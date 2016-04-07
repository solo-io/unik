package daemon

import (
	"encoding/json"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/stagers"
	"github.com/go-martini/martini"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxmartini"
	"github.com/pborman/uuid"
	"net/http"
	"strings"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"strconv"
	"fmt"
	"time"
	"io"
	"mime/multipart"
)

type UnikDaemon struct {
	server    *martini.ClassicMartini
	stagers   map[string]stagers.Stager
	compilers map[string]compilers.Compiler
	mode      string
}

const (
	aws = "aws"
	vsphere = "vsphere"
	vbox = "vbox"
)

func NewUnikDaemon(config config.UnikConfig) *UnikDaemon {
	stagers := make(map[string]stagers.Stager)
	compilers := make(map[string]compilers.Compiler)

	if config.Config.Providers.Aws.AwsAccessKeyID != "" {
		stagers[aws] = NewAwsStager(config)
	}
	if config.Config.Providers.Vsphere.VsphereURL != "" {
		stagers[vsphere] = NewVsphereStager(config)
	}
	stagers[vbox] = NewVirtualboxStager(config)

	return &UnikDaemon{
		server: lxmartini.QuietMartini(),
		stagers: stagers,
		compilers: compilers,
		mode: vbox,
	}
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

	//mode
	d.server.Get("/modes/current", func() string {
		return d.mode
	})
	d.server.Get("/modes", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "set-staging-mode", func(logger lxlog.Logger) (interface{}, error) {
			modes := []string{}
			for mode := range d.compilers {
				modes = append(modes, mode)
			}
			return strings.Join(modes, "|"), nil
		})
	})
	d.server.Put("/mode/:name", func(res http.ResponseWriter, req *http.Request, params martini.Params){
		streamOrRespond(res, req, "set-staging-mode", func(logger lxlog.Logger) (interface{}, error) {
			mode := params["name"]
			logger.Debugf("requested staging mode: %s", mode)
			if _, ok := d.stagers[mode]; !ok {
				return nil, lxerrors.New("requested mode '"+mode+"' not found", nil)
			}
			d.mode = mode
			return mode+" accepted", nil
		})
	})

	//images
	d.server.Get("/images", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "get-images", func(logger lxlog.Logger) (interface{}, error) {
			images, err := d.stagers[d.mode].ListImages(logger)
			if err != nil {
				logger.WithErr(err).Errorf("could not get image list")
			} else {
				logger.WithFields(lxlog.Fields{
					"images": images,
				}).Debugf("Listing all images")
			}
			return images, err
		})
	})
	d.server.Get("/images/:image_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "get-images", func(logger lxlog.Logger) (interface{}, error) {
			imageName := params["image_name"]
			image, err := d.stagers[d.mode].GetImage(logger, imageName)
			if err != nil {
				logger.WithErr(err).Errorf("could not get image %s", imageName)
			} else {
				logger.WithFields(lxlog.Fields{
					"image": image,
				}).Debugf("Retrieving image")
			}
			return image, err
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
			compilerMode := getCompilerMode(d.mode, unikernelType)
			compiler, ok := d.compilers[compilerMode]
			if !ok {
				return nil, lxerrors.New("unikernel type "+unikernelType+" not available for "+d.mode+"infrastructure", nil)
			}
			mountPoints := strings.Split(req.FormValue("mounts"), ",")

			logger.WithFields(lxlog.Fields{
				"source-tar": header.Filename,
				"force":         force,
				"mount-points":   mountPoints,
				"name": name,
			}).Debugf("compiling raw image")

			rawImage, err := compiler.CompileRawImage(sourceTar, header, mountPoints)
			if err != nil {
				return nil, lxerrors.New("failed to compile raw image", err)
			}

			logger.WithFields(lxlog.Fields{
				"raw-image": rawImage,
			}).Debugf("staging compiled image")

			image, err := d.stagers[d.mode].Stage(logger, name, rawImage, force)
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
			err := d.stagers[d.mode].DeleteImage(logger, imageName, force)
			if err != nil {
				logger.WithErr(err).Errorf("could not delete image " + imageName)
				return nil, err
			}
			return nil, nil
		})
	})

	//Instances
	d.server.Get("/instances", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "get-instances", func(logger lxlog.Logger) (interface{}, error) {
			instances, err := d.stagers[d.mode].ListInstances(logger)
			if err != nil {
				logger.WithErr(err).Errorf("could not get instance list")
			} else {
				logger.WithFields(lxlog.Fields{
					"instances": instances,
				}).Debugf("Listing all instances")
			}
			return instances, err
		})
	})
	d.server.Get("/instances/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "get-instances", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			instance, err := d.stagers[d.mode].GetInstance(logger, instanceId)
			if err != nil {
				logger.WithErr(err).Errorf("could not get instance")
			} else {
				logger.WithFields(lxlog.Fields{
					"instance": instance,
				}).Debugf("Retrieved instance")
			}
			return instance, err
		})
	})
	d.server.Delete("/instances/:instance_id", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "delete-instance", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			logger.WithFields(lxlog.Fields{
				"request": req,
			}).Infof("deleting instance " + instanceId)
			err := d.stagers[d.mode].DeleteInstance(logger, instanceId)
			if err != nil {
				return nil, lxerrors.New("could not delete instance " + instanceId, err)
			}
			return nil, nil
		})
	})
	d.server.Get("/instances/:instance_id/logs", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "get-instance-logs", func(logger lxlog.Logger) (interface{}, error) {
			instanceId := params["instance_id"]
			follow := req.URL.Query().Get("follow")
			res.Write([]byte("getting logs for " + instanceId + "...\n"))
			if strings.ToLower(follow) == "true" {
				if f, ok := res.(http.Flusher); ok {
					f.Flush()
				} else {
					return nil, lxerrors.New("not a flusher", nil)
				}

				deleteOnDisconnect := req.URL.Query().Get("delete")
				if strings.ToLower(deleteOnDisconnect) == "true" {
					defer d.stagers[d.mode].DeleteInstance(logger, instanceId)
				}

				output := ioutils.NewWriteFlusher(res)
				logFn := func() (string, error) {
					return d.stagers[d.mode].GetLogs(logger, instanceId)
				}
				err := streamOutput(logFn(), output)
				if err != nil {
					logger.WithErr(err).WithFields(lxlog.Fields{
						"instanceId": instanceId,
					}).Warnf("streaming logs stopped")
				}
				return nil, nil
			}

			logs, err := d.stagers[d.mode].GetLogs(logger, instanceId)
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

			instance, err := d.stagers[d.mode].RunInstance(logger, instanceName, imageName, mntPointsToVolumeIds, env)
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
			err := d.stagers[d.mode].StartInstance(logger, instanceId)
			if err != nil {
				return nil, lxerrors.New("could not start instance " + instanceId, err)
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
			err := d.stagers[d.mode].StopInstance(logger, instanceId)
			if err != nil {
				return nil, lxerrors.New("could not stop instance " + instanceId, err)
			}
			return nil, nil
		})
	})

	//Volumes
	d.server.Get("/volumes", func(res http.ResponseWriter, req *http.Request) {
		streamOrRespond(res, req, "get-volumes", func(logger lxlog.Logger) (interface{}, error) {
			logger.Debugf("listing volumes started")
			volumes, err := d.stagers[d.mode].ListVolumes(logger)
			if err != nil {
				return nil, lxerrors.New("could not retrieve volumes", err)
			}
			logger.WithFields(lxlog.Fields{
				"volumes": volumes,
			}).Infof("volumes")
			return volumes, nil
		})
	})
	d.server.Get("/volumes/:volume_name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "delete-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			volume, err := d.stagers[d.mode].GetVolume(logger, volumeName)
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
			if sizeStr != "" {
				size, err := strconv.Atoi(sizeStr)
				if err != nil {
					return nil, lxerrors.New("could not parse given size", err)
				}
				logger.WithFields(lxlog.Fields{
					"size": size,
					"name": volumeName,
				}).Debugf("creating empty volume started")
				volume, err := d.stagers[d.mode].CreateEmptyVolume(logger, volumeName, size)
				if err != nil {
					return nil, lxerrors.New("creating volume", err)
				}
				logger.WithFields(lxlog.Fields{
					"volume": volume,
				}).Infof("volume created")
				return volume
			}

			logger.WithFields(lxlog.Fields{
				"form": req.Form,
			}).Debugf("seeking form file marked 'tarfile'")
			dataTar, header, err := req.FormFile("tarfile")
			if err != nil {
				return nil, lxerrors.New("failed to retrieve file for field 'tarfile'", err)
			}
			defer dataTar.Close()

			logger.WithFields(lxlog.Fields{
				"tarred-data": header.Filename,
				"name": volumeName,
			}).Debugf("creating volume started")
			volume, err := d.stagers[d.mode].CreateVolume(logger, volumeName, dataTar, header)
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
			forceStr := req.URL.Query().Get("force")
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}

			logger.WithFields(lxlog.Fields{
				"force": force, "name": volumeName,
			}).Debugf("deleting volume started")
			err := d.stagers[d.mode].DeleteVolume(logger, volumeName, force)
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
			instanceId := params["instance_id"]
			device := req.URL.Query().Get("device")
			if device == "" {
				return nil, lxerrors.New("must provide a device name in URL query", nil)
			}
			logger.WithFields(lxlog.Fields{
				"instance": instanceId,
				"volume":   volumeName,
			}).Debugf("attaching volume to instance")
			err := d.stagers[d.mode].AttachVolume(logger, volumeName, instanceId, device)
			if err != nil {
				return nil, lxerrors.New("could not attach volume to instance", err)
			}
			logger.WithFields(lxlog.Fields{
				"instance": instanceId,
				"volume":   volumeName,
			}).Infof("volume attached")
			return volumeName, nil
		})
	})
	d.server.Post("/volumes/:volume_name/detach", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		streamOrRespond(res, req, "detach-volume", func(logger lxlog.Logger) (interface{}, error) {
			volumeName := params["volume_name"]
			forceStr := req.URL.Query().Get("force")
			force := false
			if strings.ToLower(forceStr) == "true" {
				force = true
			}
			logger.WithFields(lxlog.Fields{
				"volume": volumeName,
			}).Debugf("detaching volume from any instance")
			err := d.stagers[d.mode].DetachVolume(logger, volumeName, force)
			if err != nil {
				return nil, lxerrors.New("could not attach volume to instance", err)
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