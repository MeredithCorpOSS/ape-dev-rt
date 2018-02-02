//
// gomanta - Go library to interact with Joyent Manta
//
//
// Copyright (c) 2016 Joyent Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Written by Daniele Stroppa <daniele.stroppa@joyent.com>
//

package gomanta

import (
	gc "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type GoMantaTestSuite struct {
}

var _ = gc.Suite(&GoMantaTestSuite{})
