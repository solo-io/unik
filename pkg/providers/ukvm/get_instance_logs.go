package ukvm

import "io/ioutil"

func (p *UkvmProvider) GetInstanceLogs(id string) (string, error) {
	instanceLogName := getInstanceLogName(id)
	logdata, err := ioutil.ReadFile(instanceLogName)
	if err != nil {
		return "", err
	}

	return string(logdata), nil
}
