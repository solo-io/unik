package photon

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func getMemMb(flavor *photon.Flavor) float64 {
	for _, quotaItem := range flavor.Cost {
		if quotaItem.Key == "vm.memory" {
			machineMem := quotaItem.Value
			switch quotaItem.Unit {
			case "GB":
				machineMem *= 1024
			case "MB":
				machineMem *= 1
			case "KB":
				machineMem /= 1024
			default:
				logrus.WithFields(logrus.Fields{"flavor": flavor.Name, "quotaItem": quotaItem}).Infof("unknown unit for mem")
				return -1
			}
		}
	}
	logrus.WithField("flavor", flavor.Name).Infof("no vm.memory found")

	return -1

}

func (p *PhotonProvider) getFlavor(image *types.Image) string {
	options := &photon.FlavorGetOptions{
		Kind: "vm",
		Name: "",
	}
	flavorList, err := p.client.Flavors.GetAll(options)
	if err != nil {
		return ""
	}
	var minFlavorIndex int = -1
	for i := range flavorList.Items {
		machineMem := getMemMb(&flavorList.Items[i])

		if machineMem >= (float64)(image.RunSpec.DefaultInstanceMemory) {
			if minFlavorIndex == -1 {
				minFlavorIndex = i
			} else {
				if machineMem < getMemMb(&flavorList.Items[minFlavorIndex]) {
					minFlavorIndex = i
				}

			}

		}
	}

	if minFlavorIndex == -1 {
		return ""
	}

	return flavorList.Items[minFlavorIndex].Name
}

func (p *PhotonProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, errors.New("instance with name "+params.Name+" already exists. virtualbox provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, errors.New("invalid mapping for volume", err)
	}

	flavor := p.getFlavor(image)
	if flavor == "" {
		return nil, errors.New("Can't get flavor for vm", nil)
	}

	vmspec := &photon.VmCreateSpec{
		Flavor:        flavor,
		SourceImageID: image.InfrastructureId,
		Name:          params.Name,
		Affinities:    nil,
		AttachedDisks: nil,
		Environment:   params.Env,
	}
	p.client.Projects.CreateVM(p.projectId, vmspec)

	return nil, errors.New("not implemented", nil)
}
