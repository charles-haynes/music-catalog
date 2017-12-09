// Copyright © 2017 Charles Haynes <ceh+github@ceh.bz>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "music-catalog",
	Args:  cobra.ExactArgs(1),
	Short: "Generate a catalog of music in a directory",
	Long: `Given a directory tree containing music files with metadata
understood by taglib (http://taglib.org/) That is ID3v1 and ID3v2 for
MP3 files, Ogg Vorbis comments and ID3 tags and Vorbis comments in
FLAC, MPC, Speex, WavPack, TrueAudio, WAV, AIFF, MP4 and ASF files.
It generates a csv catalog containing Artist Name, Album Name, and Format
on stdout.

Usage:

music-catalog <directory>`,
	Run: doCatalog,
}

type artistAlbum struct {
	Artist string
	Album  string
}

type fileTypes map[tag.FileType]interface{}
type artists map[string]fileTypes

var catalog = map[string]artists{}

func visit(path string, f os.FileInfo, err error) error {
	if f == nil || f.IsDir() {
		return nil
	}
	p, err := os.Open(path)
	if err != nil {
		// fmt.Printf("opening %s: %s\n", path, err)
		return nil
	}
	m, err := tag.ReadFrom(p)
	if err != nil {
		// fmt.Printf("reading tags %s: %s\n", path, err)
		return nil
	}
	albumName := m.Album()
	if _, ok := catalog[albumName]; !ok {
		catalog[albumName] = artists{}
	}
	artistName := m.AlbumArtist()
	if artistName == "" {
		artistName = m.Artist()
	}
	if _, ok := catalog[albumName][artistName]; !ok {
		catalog[albumName][artistName] = fileTypes{}
	}
	catalog[albumName][artistName][m.FileType()] = nil
	return nil
}

func doCatalog(cmd *cobra.Command, args []string) {
	if err := filepath.Walk(args[0], visit); err != nil {
		log.Printf("Error reading files: %v", err)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Printf("root: %s", err)
	}
	for albumName, album := range catalog {
		for artistName, artist := range album {
			fmt.Printf(`"%s","%s"`, artistName, albumName)
			for fileType := range artist {
				fmt.Printf(`,"%s"`, fileType)
			}
			fmt.Println()
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.music-catalog.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".music-catalog" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".music-catalog")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
