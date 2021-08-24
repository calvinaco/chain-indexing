package parser_test

import (
	"github.com/crypto-com/chain-indexing/usecase/parser/utils"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/crypto-com/chain-indexing/entity/command"
	command_usecase "github.com/crypto-com/chain-indexing/usecase/command"
	"github.com/crypto-com/chain-indexing/usecase/event"
	"github.com/crypto-com/chain-indexing/usecase/model"
	"github.com/crypto-com/chain-indexing/usecase/parser"
	usecase_parser_test "github.com/crypto-com/chain-indexing/usecase/parser/test"
)

var _ = Describe("ParseMsgCommands", func() {
	Describe("MsgNFTBurnNFT", func() {
		It("should parse command with effective height in the transaction", func() {
			txDecoder := utils.NewTxDecoder()
			block, _ := mustParseBlockResp(usecase_parser_test.TX_MSG_NFT_BURN_NFT_BLOCK_RESP)
			blockResults := mustParseBlockResultsResp(
				usecase_parser_test.TX_MSG_NFT_BURN_NFT_BLOCK_RESULTS_RESP,
			)
			accountAddressPrefix := "cro"
			bondingDenom := "basecro"

			anyVersion := version.Must(version.NewVersion("v0.43"))

			cmds, err := parser.ParseBlockResultsTxsMsgToCommands(
				txDecoder,
				block,
				blockResults,
				accountAddressPrefix,
				bondingDenom,
				anyVersion,
			)
			Expect(err).To(BeNil())
			Expect(cmds).To(HaveLen(1))

			Expect(cmds).To(Equal([]command.Command{
				command_usecase.NewCreateMsgNFTBurnNFT(
					event.MsgCommonParams{
						BlockHeight: int64(17699),
						TxHash:      "63B42F5AC39D758E5590E7D54A0F811968D1C5C0420EA5162CE83CA6D6818AA5",
						TxSuccess:   true,
						MsgIndex:    0,
					},
					model.MsgNFTBurnNFTParams{
						DenomId: "denomid",
						TokenId: "tokenid4",
						Sender:  "cro1nk4rq3q46ltgjghxz80hy385p9uj0tf58apkcd",
					},
				),
			}))
		})
	})
})
