package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wowsims/tbc/cmd/wowsimcli/cmd/upgrades"
	"github.com/wowsims/tbc/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

var decodeLinkCmd = &cobra.Command{
	Use:   "decodelink [link]",
	Short: "decode wowsims link/url",
	Long:  "decode wowsims link/url",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return decodeLink(args[0])
	},
}

func decodeLink(link string) error {
	payload, isRaid, err := upgrades.ParseLinkPayload(link)
	if err != nil {
		return err
	}

	var settings goproto.Message
	if isRaid {
		settings = &proto.RaidSimSettings{}
	} else {
		settings = &proto.IndividualSimSettings{}
	}

	if err := goproto.Unmarshal(payload, settings); err != nil {
		return fmt.Errorf("cannot unmarshal raw proto: %w", err)
	}

	fmt.Println(protojson.Format(settings))
	return nil
}
