package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/config"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/go-martini/martini"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxmartini"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/emc-advanced-dev/unik/pkg/providers/aws"
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox"
	"github.com/emc-advanced-dev/unik/pkg/state"
)

type UnikDaemon struct {
	server    *martini.ClassicMartini
	providers providers.Providers	`json:"providers"`
	compilers map[string]compilers.Compiler
}

const (
	//available providers
	aws_provider = "aws"
	vsphere_provider = "vsphere"
	virtualbox_provider = "virtualbox"
)

func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error) {
	_providers := make(providers.Providers)
	_compilers := make(map[string]compilers.Compiler)

	for _, awsConfig := range config.Providers.Aws {
		logrus.Infof("Bootstrapping provider %s with config %v", aws_provider, awsConfig)
		p := aws.NewAwsProvier(awsConfig)
		s, err := state.BasicStateFromFile(aws.AwsStateFile)
		if err != nil {
			logrus.WithError(err).Warnf("failed to read aws state file at %s, creating blank aws state", aws.AwsStateFile)
			s = state.NewBasicState(aws.AwsStateFile)
		}
		p = p.WithState(s)
		_providers[aws_provider] = p
		break
	}
	for _, vsphereConfig := range config.Providers.Vsphere {
		logrus.Infof("Bootstrapping provider %s with config %v", vsphere_provider, vsphereConfig)
		p, err := vsphere.NewVsphereProvier(vsphereConfig)
		if err != nil {
			return nil, lxerrors.New("initializing vsphere provider", err)
		}
		s, err := state.BasicStateFromFile(vsphere.VsphereStateFile)
		if err != nil {
			logrus.WithError(err).Warnf("failed to read vsphere state file at %s, creating blank vsphere state", vsphere.VsphereStateFile)
			s = state.NewBasicState(vsphere.VsphereStateFile)
		}
		p = p.WithState(s)
		_providers[vsphere_provider] = p
		break
	}
	for _, virtualboxConfig := range config.Providers.Virtualbox {
		logrus.Infof("Bootstrapping provider %s with config %v", virtualbox_provider, virtualboxConfig)
		p, err := virtualbox.NewVirtualboxProvider(virtualboxConfig)
		if err != nil {
			return nil, lxerrors.New("initializing virtualbox provider", err)
		}
		s, err := state.BasicStateFromFile(virtualbox.VirtualboxStateFile)
		if err != nil {
			logrus.WithError(err).Warnf("failed to read virtualbox state file at %s, creating blank virtualbox state", virtualbox.VirtualboxStateFile)
			s = state.NewBasicState(virtualbox.VirtualboxStateFile)
		}
		p = p.WithState(s)
		_providers[virtualbox_provider] = p
		break
	}

	_compilers[compilers.RUMP_GO_AWS] = &compilers.RumpCompiler{
		DockerImage: "projectunik/compilers-rump-go-xen",
		CreateImage: compilers.CreateImageAws,
	}

	_compilers[compilers.RUMP_GO_VMWARE] = &compilers.RumpCompiler{
		DockerImage: "projectunik/compilers-rump-go-hw",
		CreateImage: compilers.CreateImageVmware,
	}

	_compilers[compilers.RUMP_GO_VIRTUALBOX] = &compilers.RumpCompiler{
		DockerImage: "projectunik/compilers-rump-go-hw",
		CreateImage: compilers.CreateImageVirtualBox,
	}

	d := &UnikDaemon{
		server:    lxmartini.QuietMartini(),
		providers: _providers,
		compilers: _compilers,
	}

	d.addEndpoints()

	return d, nil
}

func (d *UnikDaemon) Run(port int) {
	d.server.RunOnAddr(fmt.Sprintf(":%v", port))
}

func (d *UnikDaemon) addEndpoints() {
	handle := func(res http.ResponseWriter, req *http.Request, action func() (interface{}, int, error)) {
		jsonObject, statusCode, err := action()
		res.WriteHeader(statusCode)
		if err != nil {
			if err := respond(res, err); err != nil {
				logrus.WithError(err).Errorf("failed to reply to http request")
			}
			logrus.WithError(err).Errorf("error handling request")
			return
		}
		if jsonObject != nil {
			if err := respond(res, jsonObject); err != nil {
				logrus.WithError(err).Errorf("failed to reply to http request")
			}
			logrus.WithField("result", jsonObject).Debugf("request finished")
		}
	}

	//images
	d.server.Get("/images", func(res http.ResponseWriter, req *http.Request) {
		handle(res, req, func() (interface{}, int, error) {
			allImages := []*types.Image{}
			for _, provider := range d.providers {
				images, err := provider.ListImages()
				if err != nil {
					return nil, http.StatusInternalServerError, lxerrors.New("could not get image list", err)
				}
				allImages = append(allImages, images...)
			}
			logrus.WithFields(logrus.Fields{
				"images": allImages,
			}).Debugf("Listing all images")
			return allImages, http.StatusOK, nil
		})
	})
	d.server.Get("/images/:image_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			imageName := params["image_name"]
			provider, err := d.providers.ProviderForImage(imageName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			image, err := provider.GetImage(imageName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return image, http.StatusOK, nil
		})
	})
	d.server.Post("/images/:name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			name := params["name"]
			if name == "" {
				return nil, http.StatusBadRequest, lxerrors.New("image must be named", nil)
			}
			err := req.ParseMultipartForm(0)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			logrus.WithFields(logrus.Fields{
				"req": req,
			}).Debugf("parsing multipart form")
			logrus.WithFields(logrus.Fields{
				"form": req.Form,
			}).Debugf("parsing form file marked 'tarfile'")
			sourceTar, _, err := req.FormFile("tarfile")
			if err != nil {
				return nil, http.StatusBadRequest, lxerrors.New("parsing form file marked 'tarfile", err)
			}
			defer sourceTar.Close()
			forceStr := req.FormValue("force")
			var force bool
			if strings.ToLower(forceStr) == "true" {
				force = true
			}
			compilerType := req.FormValue("compiler")
			args := req.FormValue("args")
			providerType := req.FormValue("provider")
			if _, ok := d.providers[providerType]; !ok {
				return nil, http.StatusBadRequest, lxerrors.New(providerType+" is not a known provider. Available: "+strings.Join(d.providers.Keys(), "|"), nil)
			}

			compilerSupported := false
			for _, supportedCompiler := range d.providers[providerType].GetConfig().SupportedCompilers {
				if supportedCompiler == compilerType {
					compilerSupported = true
				}
			}

			if !compilerSupported {
				return nil, http.StatusBadRequest, lxerrors.New("provider "+providerType+" does not support compiler "+compilerType+"; supported compilers: "+strings.Join(d.providers[providerType].GetConfig().SupportedCompilers, "|"), nil)
			}

			compiler, ok := d.compilers[compilerType]
			if !ok {
				return nil, http.StatusBadRequest, lxerrors.New("unikernel type "+ compilerType +" not available for "+providerType+"infrastructure", nil)
			}
			mntStr := req.FormValue("mounts")

			var mountPoints []string
			if len(mntStr) > 0 {
				mountPoints = strings.Split(mntStr, ",")
			}

			logrus.WithFields(logrus.Fields{
				"force":        force,
				"mount-points": mountPoints,
				"name":         name,
				"args":         args,
				"compiler":     compilerType,
				"provider":	providerType,
			}).Debugf("compiling raw image")
			rawImage, err := compiler.CompileRawImage(sourceTar, args, mountPoints)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("failed to compile raw image", err)
			}
			defer os.Remove(rawImage.LocalImagePath)
			logrus.Debugf("raw image compiled and saved to "+rawImage.LocalImagePath)

			image, err := d.providers[providerType].Stage(name, rawImage, force)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("failed staging image", err)
			}
			return image, http.StatusCreated, nil
		})
	})
	d.server.Delete("/images/:image_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			imageName := params["image_name"]
			if imageName == "" {
				logrus.WithFields(logrus.Fields{
					"request": fmt.Sprintf("%v", req),
				}).Errorf("image must be named")
				return nil, http.StatusBadRequest, lxerrors.New("image must be named", nil)
			}
			logrus.WithFields(logrus.Fields{
				"request": req,
			}).Infof("deleting instance " + imageName)
			forceStr := req.URL.Query().Get("force")
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}
			provider, err := d.providers.ProviderForImage(imageName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			err = provider.DeleteImage(imageName, force)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return nil, http.StatusNoContent, nil
		})
	})

	//Instances
	d.server.Get("/instances", func(res http.ResponseWriter, req *http.Request) {
		handle(res, req, func() (interface{}, int, error) {
			allInstances := []*types.Instance{}
			for _, provider := range d.providers {
				instances, err := provider.ListInstances()
				if err != nil {
					return nil, http.StatusInternalServerError, lxerrors.New("could not get instance list", err)
				}
				allInstances = append(allInstances, instances...)
			}
			logrus.WithFields(logrus.Fields{
				"instances": allInstances,
			}).Debugf("Listing all instances")
			return allInstances, http.StatusOK, nil
		})
	})
	d.server.Get("/instances/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			instanceId := params["instance_id"]
			provider, err := d.providers.ProviderForInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			instance, err := provider.GetInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return instance, http.StatusOK, nil
		})
	})
	d.server.Delete("/instances/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			instanceId := params["instance_id"]
			logrus.WithFields(logrus.Fields{
				"request": req,
			}).Infof("deleting instance " + instanceId)
			provider, err := d.providers.ProviderForInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			forceStr := req.URL.Query().Get("force")
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}
			err = provider.DeleteInstance(instanceId, force)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return nil, http.StatusNoContent, nil
		})
	})
	d.server.Get("/instances/:instance_id/logs", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			instanceId := params["instance_id"]
			follow := req.URL.Query().Get("follow")
			res.Write([]byte("getting logs for " + instanceId + "...\n"))
			provider, err := d.providers.ProviderForInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			if strings.ToLower(follow) == "true" {
				if f, ok := res.(http.Flusher); ok {
					f.Flush()
				} else {
					return nil, http.StatusInternalServerError, lxerrors.New("not a flusher", nil)
				}

				deleteOnDisconnect := req.URL.Query().Get("delete")
				if strings.ToLower(deleteOnDisconnect) == "true" {
					defer provider.DeleteInstance(instanceId, true)
				}

				output := ioutils.NewWriteFlusher(res)
				logFn := func() (string, error) {
					return provider.GetInstanceLogs(instanceId)
				}
				err := streamOutput(logFn, output)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"instanceId": instanceId,
					}).Warnf("streaming logs stopped")
				}
				return nil, 0, nil
			}
			logs, err := provider.GetInstanceLogs(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("failed to perform get logs request", err)
			}
			return logs, http.StatusOK, nil
		})
	})
	d.server.Post("/instances/:name/run", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			logrus.WithFields(logrus.Fields{
				"request": req, "query": req.URL.Query(),
			}).Debugf("recieved run request")

			imageName := req.URL.Query().Get("image_name")
			if imageName == "" {
				return nil, http.StatusBadRequest, lxerrors.New("image must be named", nil)
			}

			instanceName := params["name"]

			envDelimiter := req.URL.Query().Get("useDelimiter")
			if envDelimiter == "" {
				envDelimiter = ","
			}
			envPairDelimiter := req.URL.Query().Get("usePairDelimiter")
			if envPairDelimiter == "" {
				envPairDelimiter = "="
			}

			env := make(map[string]string)
			fullEnvString := req.URL.Query().Get("env")
			if len(fullEnvString) > 0 {
				envPairs := strings.Split(fullEnvString, envDelimiter)
				for _, envPair := range envPairs {
					splitEnv := strings.Split(envPair, envPairDelimiter)
					if len(splitEnv) != 2 {
						logrus.WithFields(logrus.Fields{
							"envPair": envPair,
						}).Warnf("was given a env string with an invalid format, ignoring")
						continue
					}
					env[splitEnv[0]] = splitEnv[1]
				}
			}

			mntPointsToVolumeIds := make(map[string]string)
			fullMountsString := req.URL.Query().Get("mounts")
			if len(fullMountsString) > 0 {
				//expected format:
				//vol1:/mount1,vol2:/mount2,...
				mountVolumePairs := strings.Split(fullMountsString, ",")
				for _, mountVolumePair := range mountVolumePairs {
					mountVolumeTuple := strings.Split(mountVolumePair, ":")
					if len(mountVolumeTuple) != 2 {
						logrus.WithFields(logrus.Fields{
							"mountVolumePair": mountVolumePair,
						}).Warnf("was given a mount-volume pair string with an invalid format, ignoring")
						continue
					}
					mntPointsToVolumeIds[mountVolumeTuple[1]] = mountVolumeTuple[0]
				}
			}

			provider, err := d.providers.ProviderForImage(imageName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			instance, err := provider.RunInstance(instanceName, imageName, mntPointsToVolumeIds, env)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return instance, http.StatusCreated, nil
		})
	})
	d.server. Post("/instances/:instance_id/start", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			instanceId := params["instance_id"]
			logrus.WithFields(logrus.Fields{
				"request": req,
			}).Infof("starting instance " + instanceId)
			provider, err := d.providers.ProviderForInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			err = provider.StartInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not start instance "+instanceId, err)
			}
			return nil, http.StatusOK, nil
		})
	})
	d.server.Post("/instances/:instance_id/stop", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			instanceId := params["instance_id"]
			logrus.WithFields(logrus.Fields{
				"request": req,
			}).Infof("stopping instance " + instanceId)
			provider, err := d.providers.ProviderForInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			err = provider.StopInstance(instanceId)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not stop instance "+instanceId, err)
			}
			return nil, http.StatusOK, nil
		})
	})

	//Volumes
	d.server.Get("/volumes", func(res http.ResponseWriter, req *http.Request) {
		handle(res, req, func() (interface{}, int, error) {
			logrus.Debugf("listing volumes started")
			allVolumes := []*types.Volume{}
			for _, provider := range d.providers {
				volumes, err := provider.ListVolumes()
				if err != nil {
					return nil, http.StatusInternalServerError, lxerrors.New("could not retrieve volumes", err)
				}
				allVolumes = append(allVolumes, volumes...)
			}
			logrus.WithFields(logrus.Fields{
				"volumes": allVolumes,
			}).Infof("volumes")
			return allVolumes, http.StatusOK, nil
		})
	})
	d.server.Get("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(volumeName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			volume, err := provider.GetVolume(volumeName)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not get volume", err)
			}
			logrus.WithFields(logrus.Fields{
				"volume": volume,
			}).Infof("volume retrieved")
			return volume, http.StatusOK, nil
		})
	})
	d.server.Post("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			volumeName := params["volume_name"]
			err := req.ParseMultipartForm(0)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			logrus.WithFields(logrus.Fields{
				"req": req,
			}).Debugf("parsing multipart form")

			sizeStr := req.FormValue("size")
			size, err := strconv.Atoi(sizeStr)
			if err != nil {
				return nil, http.StatusBadRequest, lxerrors.New("could not parse given size", err)
			}

			providerType := req.FormValue("provider")
			if _, ok := d.providers[providerType]; !ok {
				return nil, http.StatusBadRequest, lxerrors.New(providerType+" is not a known provider. Available: "+strings.Join(d.providers.Keys(), "|"), nil)
			}

			logrus.WithFields(logrus.Fields{
				"form": req.Form,
			}).Debugf("seeking form file marked 'tarfile'")

			var imagePath string

			provider := d.providers[providerType]

			dataTar, header, err := req.FormFile("tarfile")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"size": size,
					"name": volumeName,
				}).WithError(err).Debugf("creating empty volume started")
				imagePath, err = unikos.BuildEmptyDataVolume(size)
				if err != nil {
					return nil, http.StatusInternalServerError, lxerrors.New("failed building raw image", err)
				}
				logrus.WithFields(logrus.Fields{
					"image": imagePath,
				}).Infof("raw image created")
			} else {
				defer dataTar.Close()
				logrus.WithFields(logrus.Fields{
					"tarred-data": header.Filename,
					"name":        volumeName,
				}).Debugf("creating volume started")
				imagePath, err = unikos.BuildRawDataImage(dataTar, size, provider.GetConfig().UsePartitionTables)
				if err != nil {
					return nil, http.StatusInternalServerError, lxerrors.New("creating raw volume image", err)
				}
			}
			defer os.RemoveAll(imagePath)

			volume, err := provider.CreateVolume(volumeName, imagePath)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not create volume", err)
			}
			logrus.WithFields(logrus.Fields{
				"volume": volume,
			}).Infof("volume created")
			return volume, http.StatusCreated, nil
		})
	})
	d.server.Delete("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(volumeName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			forceStr := req.URL.Query().Get("force")
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}

			logrus.WithFields(logrus.Fields{
				"force": force, "name": volumeName,
			}).Debugf("deleting volume started")
			err = provider.DeleteVolume(volumeName, force)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not delete volume", err)
			}
			logrus.WithFields(logrus.Fields{
				"volume": volumeName,
			}).Infof("volume deleted")
			return volumeName, http.StatusNoContent, nil
		})
	})
	d.server.Post("/volumes/:volume_name/attach/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(volumeName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			instanceId := params["instance_id"]
			mount := req.URL.Query().Get("mount")
			if mount == "" {
				return nil, http.StatusBadRequest, lxerrors.New("must provide a mount point in URL query", nil)
			}
			logrus.WithFields(logrus.Fields{
				"instance": instanceId,
				"volume":   volumeName,
				"mount":    mount,
			}).Debugf("attaching volume to instance")
			err = provider.AttachVolume(volumeName, instanceId, mount)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not attach volume to instance", err)
			}
			logrus.WithFields(logrus.Fields{
				"instance": instanceId,
				"volume":   volumeName,
				"mount":    mount,
			}).Infof("volume attached")
			return volumeName, http.StatusAccepted, nil
		})
	})
	d.server.Post("/volumes/:volume_name/detach", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		handle(res, req, func() (interface{}, int, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(volumeName)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			logrus.WithFields(logrus.Fields{
				"volume": volumeName,
			}).Debugf("detaching volume from any instance")
			err = provider.DetachVolume(volumeName)
			if err != nil {
				return nil, http.StatusInternalServerError, lxerrors.New("could not detach volume from instance", err)
			}
			logrus.WithFields(logrus.Fields{
				"volume": volumeName,
			}).Infof("volume detached")
			return volumeName, http.StatusAccepted, nil
		})
	})

	//info
	d.server.Get("/available_compilers", func(res http.ResponseWriter, req *http.Request) {
		handle(res, req, func() (interface{}, int, error) {
			logrus.Debugf("listing available compilers")
			availableCompilers := []string{}
			for compilerName := range d.compilers {
				availableCompilers = append(availableCompilers, compilerName)
			}
			logrus.WithFields(logrus.Fields{
				"compilers": availableCompilers,
			}).Infof("compilers")
			return availableCompilers, http.StatusOK, nil
		})
	})
	d.server.Get("/available_providers", func(res http.ResponseWriter, req *http.Request) {
		handle(res, req, func() (interface{}, int, error) {
			logrus.Debugf("listing available providers")
			availableProviders := []string{}
			for compilerName := range d.providers {
				availableProviders = append(availableProviders, compilerName)
			}
			logrus.WithFields(logrus.Fields{
				"providers": availableProviders,
			}).Infof("providers")
			return availableProviders, http.StatusOK, nil
		})
	})
}

func streamOutput(outputFunc func() (string, error), w io.Writer) error {
	linesCounted := -1
	for {
		time.Sleep(100 * time.Millisecond)
		output, err := outputFunc()
		if err != nil {
			return lxerrors.New("could not read output", err)
		}
		logLines := strings.Split(output, "\n")
		for i, _ := range logLines {
			if linesCounted < len(logLines) && linesCounted < i {
				linesCounted = i

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				} else {
					return lxerrors.New("w is not a flusher", nil)
				}

				_, err = w.Write([]byte(logLines[linesCounted] + "\n"))
				if err != nil {
					return nil
				}
			}
		}
		_, err = w.Write([]byte{0}) //ignore errors; close comes from external
		if err != nil {
			return nil
		}
		if len(logLines)-1 == linesCounted {
			time.Sleep(2500 * time.Millisecond)
			continue
		}
	}
}

func respond(res http.ResponseWriter, message interface{}) error {
	switch message.(type) {
	case string:
		messageString := message.(string)
		data := []byte(messageString)
		_, err := res.Write(data)
		if err != nil {
			return lxerrors.New("writing data", err)
		}
		return nil
	case error:
		responseError := message.(error)
		_, err := res.Write([]byte(responseError.Error()))
		if err != nil {
			return lxerrors.New("writing data", err)
		}
		return nil
	}
	data, err := json.Marshal(message)
	if err != nil {
		return lxerrors.New("marshalling message to json", err)
	}
	_, err = res.Write(data)
	if err != nil {
		return lxerrors.New("writing data", err)
	}
	return nil
}
