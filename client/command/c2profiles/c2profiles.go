package c2profiles

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

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
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/bishopfox/sliver/client/command/settings"
	"github.com/bishopfox/sliver/client/console"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/server/configs"
	"github.com/bishopfox/sliver/server/db/models"
)

// C2ProfileCmd list available http profiles
func C2ProfileCmd(cmd *cobra.Command, con *console.SliverConsoleClient, args []string) {
	profileName, _ := cmd.Flags().GetString("name")

	profile, err := con.Rpc.GetHTTPC2ProfileByName(context.Background(), &clientpb.C2ProfileReq{Name: profileName})
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	PrintC2Profiles(profile, con)
}

func ImportC2ProfileCmd(cmd *cobra.Command, con *console.SliverConsoleClient, args []string) {
	profileName, _ := cmd.Flags().GetString("name")
	filepath, _ := cmd.Flags().GetString("file")

	// retrieve and unmarshal profile config
	jsonFile, err := os.Open(filepath)
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	byteFile, _ := io.ReadAll(jsonFile)
	var config *configs.HTTPC2Config = &configs.HTTPC2Config{}
	json.Unmarshal(byteFile, config)
	_, err = con.Rpc.SaveHTTPC2Profile(context.Background(), C2ConfigToProtobuf(profileName, config))
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
}

// convert json to protobuf
func C2ConfigToProtobuf(profileName string, config *configs.HTTPC2Config) *clientpb.HTTPC2Config {

	httpC2UrlParameters := []*clientpb.HTTPC2URLParameter{}
	httpC2Headers := []*clientpb.HTTPC2Header{}
	pathSegments := []*clientpb.HTTPC2PathSegment{}

	// files
	for _, poll := range config.ImplantConfig.PollFiles {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      true,
			SegmentType: 0,
			Value:       poll,
		})
	}

	for _, session := range config.ImplantConfig.SessionFiles {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      true,
			SegmentType: 1,
			Value:       session,
		})
	}

	for _, close := range config.ImplantConfig.CloseFiles {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      true,
			SegmentType: 2,
			Value:       close,
		})
	}

	for _, stager := range config.ImplantConfig.StagerFiles {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      true,
			SegmentType: 3,
			Value:       stager,
		})
	}

	// paths
	for _, poll := range config.ImplantConfig.PollPaths {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      false,
			SegmentType: 0,
			Value:       poll,
		})
	}

	for _, session := range config.ImplantConfig.SessionPaths {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      false,
			SegmentType: 1,
			Value:       session,
		})
	}

	for _, close := range config.ImplantConfig.ClosePaths {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      false,
			SegmentType: 2,
			Value:       close,
		})
	}

	for _, stager := range config.ImplantConfig.StagerPaths {
		pathSegments = append(pathSegments, &clientpb.HTTPC2PathSegment{
			IsFile:      false,
			SegmentType: 3,
			Value:       stager,
		})
	}

	implantConfig := &clientpb.HTTPC2ImplantConfig{
		UserAgent:                 config.ImplantConfig.UserAgent,
		ChromeBaseVersion:         int32(config.ImplantConfig.ChromeBaseVersion),
		MacOSVersion:              config.ImplantConfig.MacOSVersion,
		NonceQueryArgChars:        config.ImplantConfig.NonceQueryArgChars,
		ExtraURLParameters:        httpC2UrlParameters,
		Headers:                   httpC2Headers,
		MaxFiles:                  int32(config.ImplantConfig.MaxFiles),
		MinFiles:                  int32(config.ImplantConfig.MinFiles),
		MaxPaths:                  int32(config.ImplantConfig.MaxPaths),
		MinPaths:                  int32(config.ImplantConfig.MinFiles),
		StagerFileExtension:       config.ImplantConfig.StagerFileExt,
		PollFileExtension:         config.ImplantConfig.PollFileExt,
		StartSessionFileExtension: config.ImplantConfig.StartSessionFileExt,
		SessionFileExtension:      config.ImplantConfig.SessionFileExt,
		CloseFileExtension:        config.ImplantConfig.CloseFileExt,
		PathSegments:              pathSegments,
	}

	// Server Config
	serverHeaders := []*clientpb.HTTPC2Header{}
	for _, serverHeader := range config.ServerConfig.Headers {
		for _, method := range serverHeader.Methods {
			serverHeaders = append(serverHeaders, &clientpb.HTTPC2Header{
				Method:      method,
				Name:        serverHeader.Name,
				Value:       serverHeader.Value,
				Probability: int32(serverHeader.Probability),
			})
		}
	}

	serverCookies := []*clientpb.HTTPC2Cookie{}
	for _, cookie := range config.ServerConfig.Cookies {
		serverCookies = append(serverCookies, &clientpb.HTTPC2Cookie{
			Name: cookie,
		})
	}
	serverConfig := &clientpb.HTTPC2ServerConfig{
		RandomVersionHeaders: config.ServerConfig.RandomVersionHeaders,
		Headers:              serverHeaders,
		Cookies:              serverCookies,
	}

	return &clientpb.HTTPC2Config{
		Name:          profileName,
		ImplantConfig: implantConfig,
		ServerConfig:  serverConfig,
	}
}

// PrintImplantBuilds - Print the implant builds on the server
func PrintC2Profiles(profile *clientpb.HTTPC2Config, con *console.SliverConsoleClient) {
	profileModel := models.HTTPC2ConfigFromProtobuf(profile)

	tw := table.NewWriter()
	tw.SetStyle(settings.GetTableStyle(con))
	tw.AppendHeader(table.Row{
		"Parameter",
		"Value",
	})

	// Profile metadata

	tw.AppendRow(table.Row{
		"Profile Name",
		profileModel.Name,
	})

	// Server side configuration

	var serverHeaders []string
	for _, header := range profileModel.ServerConfig.Headers {
		serverHeaders = append(serverHeaders, header.Value)
	}
	tw.AppendRow(table.Row{
		"Server Headers",
		strings.Join(serverHeaders[:], ","),
	})

	var serverCookies []string
	for _, cookie := range profileModel.ServerConfig.Cookies {
		serverCookies = append(serverCookies, cookie.Name)
	}
	tw.AppendRow(table.Row{
		"Server Cookies",
		strings.Join(serverCookies[:], ","),
	})

	tw.AppendRow(table.Row{
		"Randomize Server Headers",
		profileModel.ServerConfig.RandomVersionHeaders,
	})

	// Client side configuration

	var clientHeaders []string
	for _, header := range profileModel.ImplantConfig.Headers {
		clientHeaders = append(clientHeaders, header.Value)
	}
	tw.AppendRow(table.Row{
		"Client Headers",
		strings.Join(clientHeaders[:], ","),
	})

	var clientUrlParams []string
	for _, clientUrlParam := range profileModel.ImplantConfig.ExtraURLParameters {
		clientUrlParams = append(clientUrlParams, clientUrlParam.Name)
	}
	tw.AppendRow(table.Row{
		"Extra URL Parameters",
		strings.Join(clientUrlParams[:], ","),
	})
	tw.AppendRow(table.Row{
		"User agent",
		profileModel.ImplantConfig.UserAgent,
	})
	tw.AppendRow(table.Row{
		"Chrome base version",
		profileModel.ImplantConfig.ChromeBaseVersion,
	})
	tw.AppendRow(table.Row{
		"MacOS version",
		profileModel.ImplantConfig.MacOSVersion,
	})
	tw.AppendRow(table.Row{
		"Nonce query arg chars",
		profileModel.ImplantConfig.NonceQueryArgChars,
	})
	tw.AppendRow(table.Row{
		"Max files",
		profileModel.ImplantConfig.MaxFiles,
	})
	tw.AppendRow(table.Row{
		"Min files",
		profileModel.ImplantConfig.MinFiles,
	})
	tw.AppendRow(table.Row{
		"Max paths",
		profileModel.ImplantConfig.MaxPaths,
	})
	tw.AppendRow(table.Row{
		"Min paths",
		profileModel.ImplantConfig.MinPaths,
	})

	tw.AppendRow(table.Row{
		"Stager file extension",
		profileModel.ImplantConfig.StagerFileExtension,
	})
	tw.AppendRow(table.Row{
		"Start session file extension",
		profileModel.ImplantConfig.StartSessionFileExtension,
	})
	tw.AppendRow(table.Row{
		"Session file extension",
		profileModel.ImplantConfig.SessionFileExtension,
	})
	tw.AppendRow(table.Row{
		"Close file extension",
		profileModel.ImplantConfig.CloseFileExtension,
	})

	var (
		pollPaths    []string
		pollFiles    []string
		sessionPaths []string
		sessionFiles []string
		closePaths   []string
		closeFiles   []string
	)
	for _, segment := range profileModel.ImplantConfig.PathSegments {
		if segment.IsFile {
			switch segment.SegmentType {
			case 0:
				pollFiles = append(pollFiles, segment.Value)
			case 1:
				sessionFiles = append(sessionFiles, segment.Value)
			case 2:
				closeFiles = append(closeFiles, segment.Value)
			}
		} else {
			switch segment.SegmentType {
			case 0:
				pollPaths = append(pollPaths, segment.Value)
			case 1:
				sessionPaths = append(sessionPaths, segment.Value)
			case 2:
				closePaths = append(closePaths, segment.Value)
			}
		}
	}
	tw.AppendRow(table.Row{
		"Poll paths",
		strings.Join(pollPaths[:], ","),
	})
	tw.AppendRow(table.Row{
		"Poll files",
		strings.Join(pollFiles[:], ","),
	})
	tw.AppendRow(table.Row{
		"Session paths",
		strings.Join(sessionPaths[:], ","),
	})
	tw.AppendRow(table.Row{
		"Session files",
		strings.Join(sessionFiles[:], ","),
	})
	tw.AppendRow(table.Row{
		"Close paths",
		strings.Join(closePaths[:], ","),
	})
	tw.AppendRow(table.Row{
		"Close files",
		strings.Join(closeFiles[:], ","),
	})

	con.Println(tw.Render())
	con.Println("\n")
}
