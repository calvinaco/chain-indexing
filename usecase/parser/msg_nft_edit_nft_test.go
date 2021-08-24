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
	Describe("MsgNFTEditNFT", func() {
		It("should parse command with effective height in the transaction", func() {
			txDecoder := utils.NewTxDecoder()
			block, _ := mustParseBlockResp(usecase_parser_test.TX_MSG_NFT_EDIT_NFT_BLOCK_RESP)
			blockResults := mustParseBlockResultsResp(
				usecase_parser_test.TX_MSG_NFT_EDIT_NFT_BLOCK_RESULTS_RESP,
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
				command_usecase.NewCreateMsgNFTEditNFT(
					event.MsgCommonParams{
						BlockHeight: int64(16754),
						TxHash:      "652358F0C39307A88C347013DF65317498DB81FE847821066559E401AFF8F632",
						TxSuccess:   true,
						MsgIndex:    0,
					},
					model.MsgNFTEditNFTParams{
						DenomId:   "denomid",
						TokenId:   "tokenid8",
						TokenName: "newName",
						URI:       "newUri",
						Data:      "newData",
						Sender:    "cro1nk4rq3q46ltgjghxz80hy385p9uj0tf58apkcd",
					},
				),
			}))
		})
	})
})
