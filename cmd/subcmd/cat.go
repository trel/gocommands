package subcmd

import (
	"fmt"
	"io"
	"os"

	irodsclient_fs "github.com/cyverse/go-irodsclient/fs"
	"github.com/cyverse/gocommands/commons"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat [data-object]",
	Short: "Display the content of an iRODS data-object",
	Long:  `This displays the content of an iRODS data-object.`,
	RunE:  processCatCommand,
}

func AddCatCommand(rootCmd *cobra.Command) {
	// attach common flags
	commons.SetCommonFlags(catCmd)

	rootCmd.AddCommand(catCmd)
}

func processCatCommand(command *cobra.Command, args []string) error {
	logger := log.WithFields(log.Fields{
		"package":  "main",
		"function": "processCatCommand",
	})

	cont, err := commons.ProcessCommonFlags(command)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}

	if !cont {
		return nil
	}

	// handle local flags
	_, err = commons.InputMissingFields()
	if err != nil {
		logger.Error(err)
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}

	// Create a file system
	account := commons.GetAccount()
	filesystem, err := commons.GetIRODSFSClient(account)
	if err != nil {
		logger.Error(err)
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}

	defer filesystem.Release()

	if len(args) == 0 {
		err := fmt.Errorf("not enough input arguments")
		logger.Error(err)
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}

	for _, sourcePath := range args {
		err = catOne(filesystem, sourcePath)
		if err != nil {
			logger.Error(err)
			fmt.Fprintln(os.Stderr, err.Error())
			return nil
		}
	}
	return nil
}

func catOne(filesystem *irodsclient_fs.FileSystem, targetPath string) error {
	logger := log.WithFields(log.Fields{
		"package":  "main",
		"function": "catOne",
	})

	cwd := commons.GetCWD()
	home := commons.GetHomeDir()
	zone := commons.GetZone()
	targetPath = commons.MakeIRODSPath(cwd, home, zone, targetPath)

	targetEntry, err := filesystem.Stat(targetPath)
	if err != nil {
		return err
	}

	if targetEntry.Type == irodsclient_fs.FileEntry {
		// file
		logger.Debugf("showing the content of a data object %s", targetPath)
		fh, err := filesystem.OpenFile(targetPath, "", "r")
		if err != nil {
			return err
		}

		defer fh.Close()

		buf := make([]byte, 10240) // 10KB buffer
		for {
			readLen, err := fh.Read(buf)
			if readLen > 0 {
				fmt.Printf("%s", string(buf[:readLen]))
			}

			if err == io.EOF {
				// EOF
				break
			}
		}

	} else {
		// dir
		return fmt.Errorf("cannot show the content of a collection")
	}
	return nil
}
