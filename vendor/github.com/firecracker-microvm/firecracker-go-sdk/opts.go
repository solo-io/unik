// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package firecracker

import (
	"github.com/sirupsen/logrus"
)

type Opt func(*Machine)

func WithClient(client Firecracker) Opt {
	return func(machine *Machine) {
		machine.client = client
	}
}

func WithLogger(logger *logrus.Entry) Opt {
	return func(machine *Machine) {
		machine.logger = logger
	}
}
