package parser_test

import (
	"regexp"
	"strings"

	"github.com/crypto-com/chain-indexing/internal/json"
	"github.com/hashicorp/go-version"

	"github.com/crypto-com/chain-indexing/usecase/parser/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/crypto-com/chain-indexing/infrastructure/tendermint"
	"github.com/crypto-com/chain-indexing/usecase/event"
	"github.com/crypto-com/chain-indexing/usecase/parser"
	usecase_parser_test "github.com/crypto-com/chain-indexing/usecase/parser/test"
)

var _ = Describe("ParseMsgCommands", func() {
	Describe("MsgIBCChannelOpenConfirm", func() {
		It("should parse Msg commands when there is MsgChannelOpenConfirm in the transaction", func() {
			expected := `{
  "name": "MsgTransferCreated",
  "version": 1,
  "height": 24,
  "uuid": "{UUID}",
  "msgName": "MsgTransfer",
  "txHash": "7A5D86E0B1A364106EE2F1B40431B15A8E1B6C4A2E09E831AB773A12F5F5A564",
  "msgIndex": 0,
  "params": {
    "sourcePort": "transfer",
    "sourceChannel": "channel-0",
    "token": {
      "denom": "basecro",
	  "amount": "1234"
    },
    "sender": "cro10snhlvkpuc4xhq82uyg5ex2eezmmf5ed5tmqsv",
    "receiver": "cro1dulwqgcdpemn8c34sjd92fxepz5p0sqpeevw7f",
    "timeoutHeight": {
      "revisionNumber": "2",
      "revisionHeight": "1023"
    },
    "timeoutTimestamp": "0",

    "packetSequence": "1",
    "destinationPort": "transfer",
    "destinationChannel": "channel-0",
    "channelOrdering": "ORDER_UNORDERED",
    "connectionId": "connection-0"
  }
}
`

			txDecoder := utils.NewTxDecoder()
			block, _, _ := tendermint.ParseBlockResp(strings.NewReader(
				usecase_parser_test.TX_MSG_TRANSFER_BLOCK_RESP,
			))
			blockResults, _ := tendermint.ParseBlockResultsResp(strings.NewReader(
				usecase_parser_test.TX_MSG_TRANSFER_BLOCK_RESULTS_RESP,
			))

			accountAddressPrefix := "cro"
			stakingDenom := "basecro"
			anyVersion := version.Must(version.NewVersion("v0.43"))
			cmds, err := parser.ParseBlockResultsTxsMsgToCommands(
				txDecoder,
				block,
				blockResults,
				accountAddressPrefix,
				stakingDenom,
				anyVersion,
			)
			Expect(err).To(BeNil())
			Expect(cmds).To(HaveLen(1))
			cmd := cmds[0]
			Expect(cmd.Name()).To(Equal("CreateMsgIBCTransferTransfer"))

			untypedEvent, _ := cmd.Exec()
			typedEvent := untypedEvent.(*event.MsgIBCTransferTransfer)

			regex, _ := regexp.Compile("\n?\r?\\s?")

			Expect(json.MustMarshalToString(typedEvent)).To(Equal(
				strings.Replace(
					regex.ReplaceAllString(expected, ""),
					"{UUID}",
					typedEvent.UUID(),
					-1,
				),
			))
		})
	})
})
