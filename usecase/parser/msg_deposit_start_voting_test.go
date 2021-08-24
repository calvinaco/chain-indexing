package parser_test

import (
	"github.com/crypto-com/chain-indexing/usecase/coin"
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
	Describe("MsgDeposit", func() {
		It("should parse gov.MsgDeposit command with effective height in the transaction", func() {
			txDecoder := utils.NewTxDecoder()
			block, _ := mustParseBlockResp(usecase_parser_test.TX_MSG_DEPOSIT_AND_START_VOTING_BLOCK_RESP)
			blockResults := mustParseBlockResultsResp(
				usecase_parser_test.TX_MSG_DEPOSIT_AND_START_VOTING_BLOCK_RESULT_RESP,
			)
			accountAddressPrefix := "tcro"
			bondingDenom := "basetcro"

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
			Expect(cmds).To(HaveLen(2))

			Expect(cmds).To(Equal([]command.Command{
				command_usecase.NewCreateMsgDeposit(
					event.MsgCommonParams{
						BlockHeight: int64(439),
						TxHash:      "3EB28276333878ABCBB0D0ACB942A6F94BC23BFFE3E972B9050509D342C7F747",
						TxSuccess:   true,
						MsgIndex:    0,
					},
					model.MsgDepositParams{
						ProposalId: "1",
						Depositor:  "cro1nk4rq3q46ltgjghxz80hy385p9uj0tf58apkcd",
						Amount:     coin.MustParseCoinsNormalized("10000basecro"),
					},
				),
				command_usecase.NewStartProposalVotingPeriod(int64(439), "1"),
			}))
		})
	})
})
