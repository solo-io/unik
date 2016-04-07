package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/go-martini/martini"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxmartini"
	"github.com/pborman/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UnikDaemon struct {
	server    *martini.ClassicMartini
	providers providers.Providers
	compilers map[string]compilers.Compiler
}

const (
	aws     = "aws"
	vsphere = "vsphere"
	vbox    = "vbox"
)

func NewAwsProvider(aws config.Aws) providers.Provider {
	return nil
}

func NewVsphereProvider(aws config.Vsphere) providers.Provider {
	return nil
}

func NewVirtualboxProvider(aws config.Vbox) providers.Provider {
	return nil
}

func NewUnikDaemon(config config.UnikConfig) *UnikDaemon {
	providers := make(providers.Providers)
	compilers := make(map[string]compilers.Compiler)
	for _, awsConfig := range config.Config.Providers.Aws {
		providers[aws] = NewAwsProvider(awsConfig)
		break
	}
	for _, vsphereConfig := range config.Config.Providers.Vsphere {
		providers[vsphere] = NewVsphereProvider(vsphereConfig)
		break
	}
	for _, vboxConfig := range config.Config.Providers.Vbox {
		providers[vbox] = NewVirtualboxProvider(vboxConfig)
		break
	}
	return &UnikDaemon{
		server:    lxmartini.QuietMartini(),
		providers: providers,
		compilers: compilers,
	}
}

func (d *UnikDaemon) Start(port int) {
	d.server.RunOnAddr(fmt.Sprintf(":%v", port))
}

func (d *UnikDaemon) registerHandlers() {
	streamOrRespond := func(res http.ResponseWriter, req *http.Request, actionName string, action func(logger lxlog.Logger) (interface{}, error)) {
		verbose := req.URL.Query().Get("verbose")
		logger := lxlog.New(actionName)
		if strings.ToLower(verbose) == "true" {
			httpOutStream := ioutils.NewWriteFlusher(res)
			uuid := uuid.New()
			logger.AddWriter(uuid, lxlog.DebugLevel, httpOutStream)
			defer logger.DeleteWriter(uuid)

			jsonObject, err := action(logger)
			if err != nil {
				lxmartini.Respond(res, err)
				logger.WithErr(err).Errorf("error performing action")
				return
			}
			if text, ok := jsonObject.(string); ok {
				_, err = httpOutStream.Write([]byte(text + "\n"))
				return
			}
			if jsonObject != nil {
				httpOutStream.Write([]byte("BEGIN_JSON_DATA\n"))
				data, err := json.Marshal(jsonObject)
				if err != nil {
					lxmartini.Respond(res, lxerrors.New("could not marshal message to json", err))
					return
				}
				data = append(data, byte('\n'))
				_, err = httpOutStream.Write(data)
				if err != nil {
					lxmartini.Respond(res, lxerrors.New("could not write data", err))
					return
				}
				return
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
		} else {
			jsonObject, err := action(logger)
			if err != nil {
				lxmartini.Respond(res, err)
				logger.WithErr(err).Errorf("error performing action")
				return
			}
			if jsonObject != nil {
				lxmartini.Respond(res, jsonObject)
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
		}
	}

	//images
	d.server.Get("/images", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "get-images", func(logger lxlog.Logger) (interface{}, error) {
			allImages := []*types.Image{}
			for _, provider := range d.providers {
				images, err := provider.ListImages(logger)
				if err != nil {
					return nil, lxerrors.New("could not get image list", err)
				}
				allImages = append(allImages, images...)
			}
			logger.WithFields(lxlog.Fields{
				"images": allImages,
			}).Debugf("Listing all images")
			return allImages, nil
		})
	})
	d.server.Get("/images/:image_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "get-images", func(logger lxlog.Logger) (interface{}, error) {
			imageName := params["image_name"]
			provider, err := d.providers.ProviderForImage(logger, imageName)
			if err != nil {
				return nil, err
			}
			image, err := provider.GetImage(logger, imageName)
			if err != nil {
				return nil, err
			}
			return image, nil
		})
	})
	d.server.Post("/images/:name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "build-image", func(logger lxlog.Logger) (interface{}, error) {
			name := params["name"]
			if name == "" {
				return nil, lxerrors.New("image must be named", nil)
			}
			err := req.ParseMultipartForm(0)
			if err != nil {
				return nil, err
			}
			logger.WithFields(lxlog.Fields{
				"req": req,
			}).Debugf("parsing multipart form")
			logger.WithFields(lxlog.Fields{
				"form": req.Form,
			}).Debugf("parsing form file marked 'tarfile'")
			sourceTar, header, err := req.FormFile("tarfile")
			if err != nil {
				return nil, err
			}
			defer sourceTar.Close()
			force := req.FormValue("force")
			unikernelType := req.FormValue("type")
			providerType := req.FormValue("provider")
			if _, ok := d.providers[providerType]; !ok {
				return nil, lxerrors.New(providerType+" is not a known provider. Available: "+strings.Join(d.providers.Keys(), "|"), nil)
			}

			compilerMode := getCompilerMode(providerType, unikernelType)
			compiler, ok := d.compilers[compilerMode]
			if !ok {
				return nil, lxerrors.New("unikernel type "+unikernelType+" not available for "+providerType+"infrastructure", nil)
			}
			mountPoints := strings.Split(req.FormValue("mounts"), ",")

			compileFunc := func() (*types.RawImage, error) {
				logger.WithFields(lxlog.Fields{
					"source-tar":   header.Filename,
					"force":        force,
					"mount-points": mountPoints,
					"name":         name,
					"compiler":     compilerMode,
				}).Debugf("compiling raw image")
				rawImage, err := compiler.CompileRawImage(sourceTar, header, mountPoints)
				if err != nil {
					return nil, lxerrors.New("failed to compile raw image", err)
				}
				return rawImage, nil
			}

			image, err := d.providers[providerType].Stage(logger, name, compileFunc, force)
			if err != nil {
				return nil, lxerrors.New("failed staging image", err)
			}
			return image, nil
		})
	})
	d.server.Delete("/images/:image_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "delete-unikernel", func(logger lxlog.Logger) (interface{}, error) {
			imageName := params["image_name"]
			if imageName == "" {
				logger.WithFields(lxlog.Fields{
					"request": fmt.Sprintf("%v", req),
				}).Errorf("image must be named")
				return nil, lxerrors.New("image must be named", nil)
			}
			forceStr := req.URL.Query().Get("force")
			logger.WithFields(lxlog.Fields{
				"request": req,
			}).Infof("deleting instance " + imageName)
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}
			provider, err := d.providers.ProviderForImage(logger, imageName)
			if err != nil {
				return nil, err
			}
			err = provider.DeleteImage(logger, imageName, force)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	})

	//Instances
	d.server.Get("/instances", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "get-instances", func(logger lxlog.Logger) (interface{}, error) {
			allInstances := []*types.Instance{}
			for _, provider := range d.providers {
				instances, err := provider.ListInstances(logger)
				if err != nil {
					logger.WithErr(err).Errorf("could not get instance list")
				} else {
					allInstances = append(allInstances, instances...)
				}
			}
			logger.WithFields(lxlog.Fields{
				"instances": allInstances,
			}).Debugf("Listing all instances")
			return allInstances, nil
		})
	})
	d.server.Get("/instances/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "get-instances", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			provider, err := d.providers.ProviderForInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			instance, err := provider.GetInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			return instance, nil
		})
	})
	d.server.Delete("/instances/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "delete-instance", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			logger.WithFields(lxlog.Fields{
				"request": req,
			}).Infof("deleting instance " + instanceId)
			provider, err := d.providers.ProviderForInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			err = provider.DeleteInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	})
	d.server.Get("/instances/:instance_id/logs", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "get-instance-logs", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			follow := req.URL.Query().Get("follow")
			res.Write([]byte("getting logs for " + instanceId + "...\n"))
			provider, err := d.providers.ProviderForInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			if strings.ToLower(follow) == "true" {
				if f, ok := res.(http.Flusher); ok {
					f.Flush()
				} else {
					return nil, lxerrors.New("not a flusher", nil)
				}

				deleteOnDisconnect := req.URL.Query().Get("delete")
				if strings.ToLower(deleteOnDisconnect) == "true" {
					defer provider.DeleteInstance(logger, instanceId)
				}

				output := ioutils.NewWriteFlusher(res)
				logFn := func() (string, error) {
					return provider.GetLogs(logger, instanceId)
				}
				err := streamOutput(logFn, output)
				if err != nil {
					logger.WithErr(err).WithFields(lxlog.Fields{
						"instanceId": instanceId,
					}).Warnf("streaming logs stopped")
				}
				return nil, nil
			}
			logs, err := provider.GetLogs(logger, instanceId)
			if err != nil {
				return nil, lxerrors.New("failed to perform get logs request", err)
			}
			return logs, nil
		})
	})
	d.server.Post("/instances/:image_name/run", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "run-instance", func(logger lxlog.Logger) (interface{}, error) {
			logger.WithFields(lxlog.Fields{
				"request": req, "query": req.URL.Query(),
			}).Debugf("recieved run request")
			imageName := params["image_name"]
			if imageName == "" {
				return nil, lxerrors.New("image must be named", nil)
			}
			instanceName := req.URL.Query().Get("name")

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
						logger.WithFields(lxlog.Fields{
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
						logger.WithFields(lxlog.Fields{
							"mountVolumePair": mountVolumePair,
						}).Warnf("was given a mount-volume pair string with an invalid format, ignoring")
						continue
					}
					mntPointsToVolumeIds[mountVolumeTuple[1]] = mountVolumeTuple[0]
				}
			}

			provider, err := d.providers.ProviderForImage(logger, imageName)
			if err != nil {
				return nil, err
			}

			instance, err := provider.RunInstance(logger, instanceName, imageName, mntPointsToVolumeIds, env)
			if err != nil {
				return nil, err
			}
			return instance, nil
		})
	})
	d.server.Put("/instances/:instance_id/start", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "start-instance", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			logger.WithFields(lxlog.Fields{
				"request": req,
			}).Infof("starting instance " + instanceId)
			provider, err := d.providers.ProviderForInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			err = provider.StartInstance(logger, instanceId)
			if err != nil {
				return nil, lxerrors.New("could not start instance "+instanceId, err)
			}
			return nil, nil
		})
	})
	d.server.Put("/instances/:instance_id/stop", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "stop-instance", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			logger.WithFields(lxlog.Fields{
				"request": req,
			}).Infof("stopping instance " + instanceId)
			provider, err := d.providers.ProviderForInstance(logger, instanceId)
			if err != nil {
				return nil, err
			}
			err = provider.StopInstance(logger, instanceId)
			if err != nil {
				return nil, lxerrors.New("could not stop instance "+instanceId, err)
			}
			return nil, nil
		})
	})

	//Volumes
	d.server.Get("/volumes", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "get-volumes", func(logger lxlog.Logger) (interface{}, error) {
			logger.Debugf("listing volumes started")
			allVolumes := []*types.Volume{}
			for _, provider := range d.providers {
				volumes, err := provider.ListVolumes(logger)
				if err != nil {
					return nil, lxerrors.New("could not retrieve volumes", err)
				}
				allVolumes = append(allVolumes, volumes...)
			}
			logger.WithFields(lxlog.Fields{
				"volumes": allVolumes,
			}).Infof("volumes")
			return allVolumes, nil
		})
	})
	d.server.Get("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "delete-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(logger, volumeName)
			if err != nil {
				return nil, err
			}
			volume, err := provider.GetVolume(logger, volumeName)
			if err != nil {
				return nil, lxerrors.New("could not get volume", err)
			}
			logger.WithFields(lxlog.Fields{
				"volume": volume,
			}).Infof("volume retrieved")
			return volume, nil
		})
	})
	d.server.Post("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "create-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			err := req.ParseMultipartForm(0)
			if err != nil {
				return nil, err
			}
			logger.WithFields(lxlog.Fields{
				"req": req,
			}).Debugf("parsing multipart form")

			sizeStr := req.FormValue("sizeStr")
			size, err := strconv.Atoi(sizeStr)
			if err != nil {
				return nil, lxerrors.New("could not parse given size", err)
			}

			providerType := req.FormValue("provider")
			if _, ok := d.providers[providerType]; !ok {
				return nil, lxerrors.New(providerType+" is not a known provider. Available: "+strings.Join(d.providers.Keys(), "|"), nil)
			}

			logger.WithFields(lxlog.Fields{
				"form": req.Form,
			}).Debugf("seeking form file marked 'tarfile'")
			dataTar, header, err := req.FormFile("tarfile")
			if err != nil {
				logger.WithFields(lxlog.Fields{
					"size": size,
					"name": volumeName,
				}).WithErr(err).Debugf("creating empty volume started")
				volume, err := d.providers[providerType].CreateEmptyVolume(logger, volumeName, size)
				if err != nil {
					return nil, lxerrors.New("creating volume", err)
				}
				logger.WithFields(lxlog.Fields{
					"volume": volume,
				}).Infof("volume created")
				return volume
			}
			defer dataTar.Close()

			logger.WithFields(lxlog.Fields{
				"tarred-data": header.Filename,
				"name":        volumeName,
			}).Debugf("creating volume started")
			volume, err := d.providers[providerType].CreateVolume(logger, volumeName, dataTar, header, size)
			if err != nil {
				return nil, lxerrors.New("could not create volume", err)
			}
			logger.WithFields(lxlog.Fields{
				"volume": volume,
			}).Infof("volume created")
			return volume, nil
		})
	})
	d.server.Delete("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "delete-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(logger, volumeName)
			if err != nil {
				return nil, err
			}
			forceStr := req.URL.Query().Get("force")
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}

			logger.WithFields(lxlog.Fields{
				"force": force, "name": volumeName,
			}).Debugf("deleting volume started")
			err = provider.DeleteVolume(logger, volumeName, force)
			if err != nil {
				return nil, lxerrors.New("could not delete volume", err)
			}
			logger.WithFields(lxlog.Fields{
				"volume": volumeName,
			}).Infof("volume deleted")
			return volumeName, nil
		})
	})
	d.server.Post("/volumes/:volume_name/attach/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "attach-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(logger, volumeName)
			if err != nil {
				return nil, err
			}
			instanceId := params["instance_id"]
			mount := req.URL.Query().Get("mount")
			if mount == "" {
				return nil, lxerrors.New("must provide a mount point in URL query", nil)
			}
			logger.WithFields(lxlog.Fields{
				"instance": instanceId,
				"volume":   volumeName,
				"mount":    mount,
			}).Debugf("attaching volume to instance")
			err = provider.AttachVolume(logger, volumeName, instanceId, mount)
			if err != nil {
				return nil, lxerrors.New("could not attach volume to instance", err)
			}
			logger.WithFields(lxlog.Fields{
				"instance": instanceId,
				"volume":   volumeName,
				"mount":    mount,
			}).Infof("volume attached")
			return volumeName, nil
		})
	})
	d.server.Post("/volumes/:volume_name/detach", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "detach-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			provider, err := d.providers.ProviderForVolume(logger, volumeName)
			if err != nil {
				return nil, err
			}
			logger.WithFields(lxlog.Fields{
				"volume": volumeName,
			}).Debugf("detaching volume from any instance")
			err = provider.DetachVolume(logger, volumeName)
			if err != nil {
				return nil, lxerrors.New("could not detach volume from instance", err)
			}
			logger.WithFields(lxlog.Fields{
				"volume": volumeName,
			}).Infof("volume detached")
			return volumeName, nil
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

func getCompilerMode(stagerMode, unikernelType string) string {
	return fmt.Sprintf("%s-%s", stagerMode, unikernelType)
}
