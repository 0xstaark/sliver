package use

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"github.com/bishopfox/sliver/client/command/beacons"
	"github.com/bishopfox/sliver/client/console"
	"github.com/desertbit/grumble"
)

// UseBeaconCmd - Change the active beacon
func UseBeaconCmd(ctx *grumble.Context, con *console.SliverConsoleClient) {
	beacon, err := beacons.SelectBeacon(con)
	if beacon != nil {
		con.ActiveTarget.Set(nil, beacon)
		con.PrintInfof("Active beacon %s (%d)\n", beacon.Name, beacon.ID)
	} else if err != nil {
		switch err {
		case beacons.ErrNoBeacons:
			con.PrintErrorf("No beacon available\n")
		case beacons.ErrNoSelection:
			con.PrintErrorf("No beacon selected\n")
		default:
			con.PrintErrorf("%s\n", err)
		}
	}
}
