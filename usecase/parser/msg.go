package parser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/go-version"

	"github.com/crypto-com/chain-indexing/projection/validator/constants"
	"github.com/crypto-com/chain-indexing/usecase/model/genesis"

	"github.com/crypto-com/chain-indexing/usecase/parser/utils"

	"github.com/crypto-com/chain-indexing/usecase/parser/ibcmsg"

	"github.com/crypto-com/chain-indexing/internal/tmcosmosutils"

	"github.com/crypto-com/chain-indexing/internal/primptr"

	"github.com/crypto-com/chain-indexing/internal/utctime"

	jsoniter "github.com/json-iterator/go"

	"github.com/crypto-com/chain-indexing/entity/command"
	"github.com/crypto-com/chain-indexing/usecase/coin"
	command_usecase "github.com/crypto-com/chain-indexing/usecase/command"
	"github.com/crypto-com/chain-indexing/usecase/event"
	"github.com/crypto-com/chain-indexing/usecase/model"
)

func ParseBlockResultsTxsMsgToCommands(
	txDecoder *utils.TxDecoder,
	block *model.Block,
	blockResults *model.BlockResults,
	accountAddressPrefix string,
	stakingDenom string,
	cosmosSDKVersion *version.Version,
) ([]command.Command, error) {
	commands := make([]command.Command, 0)

	blockHeight := block.Height
	for i, txHex := range block.Txs {
		txHash := TxHash(txHex)
		txSuccess := true
		txsResult := blockResults.TxsResults[i]

		if txsResult.Code != 0 {
			txSuccess = false
		}
		tx, err := txDecoder.Decode(txHex)
		if err != nil {
			panic(fmt.Sprintf("error decoding transaction: %v", err))
		}

		for msgIndex, msg := range tx.Body.Messages {
			msgCommonParams := event.MsgCommonParams{
				BlockHeight: blockHeight,
				TxHash:      txHash,
				TxSuccess:   txSuccess,
				MsgIndex:    msgIndex,
			}

			var msgCommands []command.Command
			switch msg["@type"] {
			case "/cosmos.bank.v1beta1.MsgSend":
				msgCommands = parseMsgSend(msgCommonParams, msg)
			case "/cosmos.bank.v1beta1.MsgMultiSend":
				msgCommands = parseMsgMultiSend(msgCommonParams, msg)
			case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":
				msgCommands = parseMsgSetWithdrawAddress(msgCommonParams, msg)
			case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
				msgCommands = parseMsgWithdrawDelegatorReward(txSuccess, txsResult, msgIndex, msgCommonParams, msg)
			case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
				msgCommands = parseMsgWithdrawValidatorCommission(txSuccess, txsResult, msgIndex, msgCommonParams, msg)
			case "/cosmos.distribution.v1beta1.MsgFundCommunityPool":
				msgCommands = parseMsgFundCommunityPool(msgCommonParams, msg)
			case "/cosmos.gov.v1beta1.MsgSubmitProposal":
				msgCommands = parseMsgSubmitProposal(txSuccess, txsResult, msgIndex, msgCommonParams, msg)
			case "/cosmos.gov.v1beta1.MsgVote":
				msgCommands = parseMsgVote(msgCommonParams, msg)
			case "/cosmos.gov.v1beta1.MsgDeposit":
				msgCommands = parseMsgDeposit(msgCommonParams, txsResult, msgIndex, msg)
			case "/cosmos.staking.v1beta1.MsgDelegate":
				msgCommands = parseMsgDelegate(
					accountAddressPrefix,
					stakingDenom,
					txSuccess, txsResult, msgIndex, msgCommonParams, msg,
				)
			case "/cosmos.staking.v1beta1.MsgUndelegate":
				msgCommands = parseMsgUndelegate(
					accountAddressPrefix,
					stakingDenom,
					txSuccess, txsResult, msgIndex, msgCommonParams, msg,
				)
			case "/cosmos.staking.v1beta1.MsgBeginRedelegate":
				msgCommands = parseMsgBeginRedelegate(
					accountAddressPrefix,
					stakingDenom,
					txSuccess, txsResult, msgIndex, msgCommonParams, msg,
				)
			case "/cosmos.slashing.v1beta1.MsgUnjail":
				msgCommands = parseMsgUnjail(msgCommonParams, msg)
			case "/cosmos.staking.v1beta1.MsgCreateValidator":
				msgCommands = parseMsgCreateValidator(msgCommonParams, msg)
			case "/cosmos.staking.v1beta1.MsgEditValidator":
				msgCommands = parseMsgEditValidator(msgCommonParams, msg)
			case "/chainmain.nft.v1.MsgIssueDenom":
				msgCommands = parseMsgNFTIssueDenom(msgCommonParams, msg)
			case "/chainmain.nft.v1.MsgMintNFT":
				msgCommands = parseMsgNFTMintNFT(msgCommonParams, msg)
			case "/chainmain.nft.v1.MsgTransferNFT":
				msgCommands = parseMsgNFTTransferNFT(msgCommonParams, msg)
			case "/chainmain.nft.v1.MsgEditNFT":
				msgCommands = parseMsgNFTEditNFT(msgCommonParams, msg)
			case "/chainmain.nft.v1.MsgBurnNFT":
				msgCommands = parseMsgNFTBurnNFT(msgCommonParams, msg)
			case "/ibc.core.client.v1.MsgCreateClient":
				msgCommands = ibcmsg.ParseMsgCreateClient(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.client.v1.MsgUpdateClient":
				msgCommands = ibcmsg.ParseMsgUpdateClient(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.connection.v1.MsgConnectionOpenInit":
				msgCommands = ibcmsg.ParseMsgConnectionOpenInit(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.connection.v1.MsgConnectionOpenTry":
				msgCommands = ibcmsg.ParseMsgConnectionOpenTry(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.connection.v1.MsgConnectionOpenAck":
				msgCommands = ibcmsg.ParseMsgConnectionOpenAck(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.connection.v1.MsgConnectionOpenConfirm":
				msgCommands = ibcmsg.ParseMsgConnectionOpenConfirm(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgChannelOpenInit":
				msgCommands = ibcmsg.ParseMsgChannelOpenInit(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgChannelOpenTry":
				msgCommands = ibcmsg.ParseMsgChannelOpenTry(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgChannelOpenAck":
				msgCommands = ibcmsg.ParseMsgChannelOpenAck(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgChannelOpenConfirm":
				msgCommands = ibcmsg.ParseMsgChannelOpenConfirm(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.applications.transfer.v1.MsgTransfer":
				msgCommands = ibcmsg.ParseMsgTransfer(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgRecvPacket":
				msgCommands = ibcmsg.ParseMsgRecvPacket(msgCommonParams, txsResult, msgIndex, msg, cosmosSDKVersion)
			case "/ibc.core.channel.v1.MsgAcknowledgement":
				msgCommands = ibcmsg.ParseMsgAcknowledgement(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgTimeout":
				msgCommands = ibcmsg.ParseMsgTimeout(msgCommonParams, txsResult, msgIndex, msg)
			case "/ibc.core.channel.v1.MsgTimeoutOnClose":
				msgCommands = ibcmsg.ParseMsgTimeoutOnClose(msgCommonParams, txsResult, msgIndex, msg)
			}

			commands = append(commands, msgCommands...)
		}
	}

	return commands, nil
}

func parseMsgSend(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgSend(
		msgCommonParams,

		event.MsgSendCreatedParams{
			FromAddress: msg["from_address"].(string),
			ToAddress:   msg["to_address"].(string),
			Amount:      tmcosmosutils.MustNewCoinsFromAmountInterface(msg["amount"].([]interface{})),
		},
	)}
}

func parseMsgMultiSend(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	rawInputs, _ := msg["inputs"].([]interface{})
	inputs := make([]model.MsgMultiSendInput, 0, len(rawInputs))
	for _, rawInput := range rawInputs {
		input, _ := rawInput.(map[string]interface{})
		inputs = append(inputs, model.MsgMultiSendInput{
			Address: input["address"].(string),
			Amount:  tmcosmosutils.MustNewCoinsFromAmountInterface(input["coins"].([]interface{})),
		})
	}

	rawOutputs, _ := msg["outputs"].([]interface{})
	outputs := make([]model.MsgMultiSendOutput, 0, len(rawOutputs))
	for _, rawOutput := range rawOutputs {
		output, _ := rawOutput.(map[string]interface{})
		outputs = append(outputs, model.MsgMultiSendOutput{
			Address: output["address"].(string),
			Amount:  tmcosmosutils.MustNewCoinsFromAmountInterface(output["coins"].([]interface{})),
		})
	}

	return []command.Command{command_usecase.NewCreateMsgMultiSend(
		msgCommonParams,

		model.MsgMultiSendParams{
			Inputs:  inputs,
			Outputs: outputs,
		},
	)}
}

func parseMsgSetWithdrawAddress(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgSetWithdrawAddress(
		msgCommonParams,

		model.MsgSetWithdrawAddressParams{
			DelegatorAddress: msg["delegator_address"].(string),
			WithdrawAddress:  msg["withdraw_address"].(string),
		},
	)}
}

func parseMsgWithdrawDelegatorReward(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	if !txSuccess {
		delegatorAddress, _ := msg["delegator_address"].(string)
		return []command.Command{command_usecase.NewCreateMsgWithdrawDelegatorReward(
			msgCommonParams,

			model.MsgWithdrawDelegatorRewardParams{
				DelegatorAddress: delegatorAddress,
				ValidatorAddress: msg["validator_address"].(string),
				RecipientAddress: delegatorAddress,
				Amount:           coin.NewEmptyCoins(),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	var recipient string
	var amount coin.Coins
	// When there is no reward withdrew, `transfer` event would not exist
	if event := log.GetEventByType("transfer"); event == nil {
		recipient, _ = msg["delegator_address"].(string)
		amount = coin.NewEmptyCoins()
	} else {
		recipient = event.MustGetAttributeByKey("recipient")
		amountValue := event.MustGetAttributeByKey("amount")
		amount = coin.MustParseCoinsNormalized(amountValue)
	}

	return []command.Command{command_usecase.NewCreateMsgWithdrawDelegatorReward(
		msgCommonParams,

		model.MsgWithdrawDelegatorRewardParams{
			DelegatorAddress: msg["delegator_address"].(string),
			ValidatorAddress: msg["validator_address"].(string),
			RecipientAddress: recipient,
			Amount:           amount,
		},
	)}
}

func parseMsgWithdrawValidatorCommission(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgWithdrawValidatorCommission(
			msgCommonParams,

			model.MsgWithdrawValidatorCommissionParams{
				ValidatorAddress: msg["validator_address"].(string),
				RecipientAddress: "",
				Amount:           coin.NewEmptyCoins(),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	var recipient string
	var amount coin.Coins
	// When there is no reward withdrew, `transfer` event would not exist
	if event := log.GetEventByType("transfer"); event == nil {
		recipient, _ = msg["delegator_address"].(string)
		amount = coin.NewEmptyCoins()
	} else {
		recipient = event.MustGetAttributeByKey("recipient")
		amountValue := event.MustGetAttributeByKey("amount")
		amount = coin.MustParseCoinsNormalized(amountValue)
	}

	return []command.Command{command_usecase.NewCreateMsgWithdrawValidatorCommission(
		msgCommonParams,

		model.MsgWithdrawValidatorCommissionParams{
			ValidatorAddress: msg["validator_address"].(string),
			RecipientAddress: recipient,
			Amount:           amount,
		},
	)}
}

func parseMsgFundCommunityPool(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgFundCommunityPool(
		msgCommonParams,

		model.MsgFundCommunityPoolParams{
			Depositor: msg["depositor"].(string),
			Amount:    tmcosmosutils.MustNewCoinsFromAmountInterface(msg["amount"].([]interface{})),
		},
	)}
}

func parseMsgSubmitProposal(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	rawContent, err := jsoniter.Marshal(msg["content"])
	if err != nil {
		panic(fmt.Sprintf("error encoding proposal content: %v", err))
	}
	var proposalContent model.MsgSubmitProposalContent
	if err := jsoniter.Unmarshal(rawContent, &proposalContent); err != nil {
		panic(fmt.Sprintf("error decoding proposal content: %v", err))
	}

	var cmds []command.Command
	if proposalContent.Type == "/cosmos.params.v1beta1.ParameterChangeProposal" {
		cmds = parseMsgSubmitParamChangeProposal(txSuccess, txsResult, msgIndex, msgCommonParams, msg, rawContent)
	} else if proposalContent.Type == "/cosmos.distribution.v1beta1.CommunityPoolSpendProposal" {
		cmds = parseMsgSubmitCommunityFundSpendProposal(txSuccess, txsResult, msgIndex, msgCommonParams, msg, rawContent)
	} else if proposalContent.Type == "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal" {
		cmds = parseMsgSubmitSoftwareUpgradeProposal(txSuccess, txsResult, msgIndex, msgCommonParams, msg, rawContent)
	} else if proposalContent.Type == "/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal" {
		cmds = parseMsgSubmitCancelSoftwareUpgradeProposal(txSuccess, txsResult, msgIndex, msgCommonParams, msg, rawContent)
	} else if proposalContent.Type == "/cosmos.gov.v1beta1.TextProposal" {
		cmds = parseMsgSubmitTextProposal(txSuccess, txsResult, msgIndex, msgCommonParams, msg, rawContent)
	} else {
		panic(fmt.Sprintf("unrecognzied govenance proposal type `%s`", proposalContent.Type))
	}

	if msgCommonParams.TxSuccess {
		log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
		logEvent := log.GetEventByType("submit_proposal")
		if logEvent == nil {
			panic("missing `submit_proposal` event in TxsResult log")
		}

		if logEvent.HasAttribute("voting_period_start") {
			cmds = append(cmds, command_usecase.NewStartProposalVotingPeriod(
				msgCommonParams.BlockHeight, logEvent.MustGetAttributeByKey("voting_period_start"),
			))
		}
	}

	return cmds
}

func parseMsgSubmitParamChangeProposal(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
	rawContent []byte,
) []command.Command {
	var proposalContent model.MsgSubmitParamChangeProposalContent
	if err := jsoniter.Unmarshal(rawContent, &proposalContent); err != nil {
		panic("error decoding param change proposal content")
	}

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgSubmitParamChangeProposal(
			msgCommonParams,

			model.MsgSubmitParamChangeProposalParams{
				MaybeProposalId: nil,
				Content:         proposalContent,
				ProposerAddress: msg["proposer"].(string),
				InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
					msg["initial_deposit"].([]interface{}),
				),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	event := log.GetEventByType("submit_proposal")
	if event == nil {
		panic("missing `submit_proposal` event in TxsResult log")
	}
	proposalId := event.GetAttributeByKey("proposal_id")
	if proposalId == nil {
		panic("missing `proposal_id` in `submit_proposal` event of TxsResult log")
	}

	return []command.Command{command_usecase.NewCreateMsgSubmitParamChangeProposal(
		msgCommonParams,

		model.MsgSubmitParamChangeProposalParams{
			MaybeProposalId: proposalId,
			Content:         proposalContent,
			ProposerAddress: msg["proposer"].(string),
			InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
				msg["initial_deposit"].([]interface{}),
			),
		},
	)}
}

func parseMsgSubmitCommunityFundSpendProposal(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
	rawContent []byte,
) []command.Command {
	var rawProposalContent model.RawMsgSubmitCommunityPoolSpendProposalContent
	if err := jsoniter.Unmarshal(rawContent, &rawProposalContent); err != nil {
		panic("error decoding community pool spend proposal content")
	}
	proposalContent := model.MsgSubmitCommunityPoolSpendProposalContent{
		Type:             rawProposalContent.Type,
		Title:            rawProposalContent.Title,
		Description:      rawProposalContent.Description,
		RecipientAddress: rawProposalContent.RecipientAddress,
		Amount:           tmcosmosutils.MustNewCoinsFromAmountInterface(rawProposalContent.Amount),
	}

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgSubmitCommunityPoolSpendProposal(
			msgCommonParams,

			model.MsgSubmitCommunityPoolSpendProposalParams{
				MaybeProposalId: nil,
				Content:         proposalContent,
				ProposerAddress: msg["proposer"].(string),
				InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
					msg["initial_deposit"].([]interface{}),
				),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	// When there is no reward withdrew, `transfer` event would not exist
	event := log.GetEventByType("submit_proposal")
	if event == nil {
		panic("missing `submit_proposal` event in TxsResult log")
	}
	proposalId := event.GetAttributeByKey("proposal_id")
	if proposalId == nil {
		panic("missing `proposal_id` in `submit_proposal` event of TxsResult log")
	}

	return []command.Command{command_usecase.NewCreateMsgSubmitCommunityPoolSpendProposal(
		msgCommonParams,

		model.MsgSubmitCommunityPoolSpendProposalParams{
			MaybeProposalId: proposalId,
			Content:         proposalContent,
			ProposerAddress: msg["proposer"].(string),
			InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
				msg["initial_deposit"].([]interface{}),
			),
		},
	)}
}

func parseMsgSubmitSoftwareUpgradeProposal(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
	rawContent []byte,
) []command.Command {
	var rawProposalContent model.RawMsgSubmitSoftwareUpgradeProposalContent
	if err := jsoniter.Unmarshal(rawContent, &rawProposalContent); err != nil {
		panic("error decoding software upgrade proposal content")
	}

	height, err := strconv.ParseInt(rawProposalContent.Plan.Height, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("error parsing software upgrade proposal plan height: %v", err))
	}
	proposalContent := model.MsgSubmitSoftwareUpgradeProposalContent{
		Type:        rawProposalContent.Type,
		Title:       rawProposalContent.Title,
		Description: rawProposalContent.Description,
		Plan: model.MsgSubmitSoftwareUpgradeProposalPlan{
			Name:   rawProposalContent.Plan.Name,
			Time:   utctime.FromTime(rawProposalContent.Plan.Time),
			Height: height,
			Info:   rawProposalContent.Plan.Info,
		},
	}

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgSubmitSoftwareUpgradeProposal(
			msgCommonParams,

			model.MsgSubmitSoftwareUpgradeProposalParams{
				MaybeProposalId: nil,
				Content:         proposalContent,
				ProposerAddress: msg["proposer"].(string),
				InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
					msg["initial_deposit"].([]interface{}),
				),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	// When there is no reward withdrew, `transfer` event would not exist
	event := log.GetEventByType("submit_proposal")
	if event == nil {
		panic("missing `submit_proposal` event in TxsResult log")
	}
	proposalId := event.GetAttributeByKey("proposal_id")
	if proposalId == nil {
		panic("missing `proposal_id` in `submit_proposal` event of TxsResult log")
	}

	return []command.Command{command_usecase.NewCreateMsgSubmitSoftwareUpgradeProposal(
		msgCommonParams,

		model.MsgSubmitSoftwareUpgradeProposalParams{
			MaybeProposalId: proposalId,
			Content:         proposalContent,
			ProposerAddress: msg["proposer"].(string),
			InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
				msg["initial_deposit"].([]interface{}),
			),
		},
	)}
}

func parseMsgSubmitCancelSoftwareUpgradeProposal(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
	rawContent []byte,
) []command.Command {
	var proposalContent model.MsgSubmitCancelSoftwareUpgradeProposalContent
	if err := jsoniter.Unmarshal(rawContent, &proposalContent); err != nil {
		panic("error decoding software upgrade proposal content")
	}

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgSubmitCancelSoftwareUpgradeProposal(
			msgCommonParams,

			model.MsgSubmitCancelSoftwareUpgradeProposalParams{
				MaybeProposalId: nil,
				Content:         proposalContent,
				ProposerAddress: msg["proposer"].(string),
				InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
					msg["initial_deposit"].([]interface{}),
				),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	// When there is no reward withdrew, `transfer` event would not exist
	event := log.GetEventByType("submit_proposal")
	if event == nil {
		panic("missing `submit_proposal` event in TxsResult log")
	}
	proposalId := event.GetAttributeByKey("proposal_id")
	if proposalId == nil {
		panic("missing `proposal_id` in `submit_proposal` event of TxsResult log")
	}

	return []command.Command{command_usecase.NewCreateMsgSubmitCancelSoftwareUpgradeProposal(
		msgCommonParams,

		model.MsgSubmitCancelSoftwareUpgradeProposalParams{
			MaybeProposalId: proposalId,
			Content:         proposalContent,
			ProposerAddress: msg["proposer"].(string),
			InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
				msg["initial_deposit"].([]interface{}),
			),
		},
	)}
}

func parseMsgSubmitTextProposal(
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
	rawContent []byte,
) []command.Command {
	var proposalContent model.MsgSubmitTextProposalContent
	if err := jsoniter.Unmarshal(rawContent, &proposalContent); err != nil {
		panic("error decoding text proposal content")
	}

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgSubmitTextProposal(
			msgCommonParams,

			model.MsgSubmitTextProposalParams{
				MaybeProposalId: nil,
				Content:         proposalContent,
				ProposerAddress: msg["proposer"].(string),
				InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
					msg["initial_deposit"].([]interface{}),
				),
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	// When there is no reward withdrew, `transfer` event would not exist
	event := log.GetEventByType("submit_proposal")
	if event == nil {
		panic("missing `submit_proposal` event in TxsResult log")
	}
	proposalId := event.GetAttributeByKey("proposal_id")
	if proposalId == nil {
		panic("missing `proposal_id` in `submit_proposal` event of TxsResult log")
	}

	return []command.Command{command_usecase.NewCreateMsgSubmitTextProposal(
		msgCommonParams,

		model.MsgSubmitTextProposalParams{
			MaybeProposalId: proposalId,
			Content:         proposalContent,
			ProposerAddress: msg["proposer"].(string),
			InitialDeposit: tmcosmosutils.MustNewCoinsFromAmountInterface(
				msg["initial_deposit"].([]interface{}),
			),
		},
	)}
}

func parseMsgVote(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgVote(
		msgCommonParams,

		model.MsgVoteParams{
			ProposalId: msg["proposal_id"].(string),
			Voter:      msg["voter"].(string),
			Option:     msg["option"].(string),
		},
	)}
}

func parseMsgDeposit(
	msgCommonParams event.MsgCommonParams,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msg map[string]interface{},
) []command.Command {
	cmds := []command.Command{command_usecase.NewCreateMsgDeposit(
		msgCommonParams,

		model.MsgDepositParams{
			ProposalId: msg["proposal_id"].(string),
			Depositor:  msg["depositor"].(string),
			Amount:     tmcosmosutils.MustNewCoinsFromAmountInterface(msg["amount"].([]interface{})),
		},
	)}

	if msgCommonParams.TxSuccess {
		log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
		logEvents := log.GetEventsByType("proposal_deposit")
		if logEvents == nil {
			panic("missing `proposal_deposit` event in TxsResult log")
		}

		for _, logEvent := range logEvents {
			if logEvent.HasAttribute("voting_period_start") {
				cmds = append(cmds, command_usecase.NewStartProposalVotingPeriod(
					msgCommonParams.BlockHeight, logEvent.MustGetAttributeByKey("voting_period_start"),
				))
				break
			}
		}
	}

	return cmds
}

func parseMsgDelegate(
	addressPrefix string,
	stakingDenom string,
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	amountValue, _ := msg["amount"].(map[string]interface{})
	amount := tmcosmosutils.MustNewCoinFromAmountInterface(amountValue)

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgDelegate(
			msgCommonParams,

			model.MsgDelegateParams{
				DelegatorAddress:   msg["delegator_address"].(string),
				ValidatorAddress:   msg["validator_address"].(string),
				Amount:             amount,
				AutoClaimedRewards: coin.Coin{},
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])

	moduleAccounts := tmcosmosutils.NewModuleAccounts(addressPrefix)
	transferEvents := log.GetEventsByType("transfer")
	autoClaimedRewards := coin.NewZeroCoin(stakingDenom)
	for _, transferEvent := range transferEvents {
		sender := transferEvent.MustGetAttributeByKey("sender")
		if sender != moduleAccounts.Distribution {
			continue
		}

		amount := transferEvent.MustGetAttributeByKey("amount")
		coin, coinErr := coin.ParseCoinNormalized(amount)
		if coinErr != nil {
			panic(fmt.Errorf("error parsing auto claimed rewards amount: %v", coinErr))
		}
		autoClaimedRewards = autoClaimedRewards.Add(coin)
	}

	return []command.Command{command_usecase.NewCreateMsgDelegate(
		msgCommonParams,

		model.MsgDelegateParams{
			DelegatorAddress:   msg["delegator_address"].(string),
			ValidatorAddress:   msg["validator_address"].(string),
			Amount:             amount,
			AutoClaimedRewards: autoClaimedRewards,
		},
	)}
}

func parseMsgUndelegate(
	addressPrefix string,
	stakingDenom string,
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	amountValue, _ := msg["amount"].(map[string]interface{})
	amount := tmcosmosutils.MustNewCoinFromAmountInterface(amountValue)

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgUndelegate(
			msgCommonParams,

			model.MsgUndelegateParams{
				DelegatorAddress:      msg["delegator_address"].(string),
				ValidatorAddress:      msg["validator_address"].(string),
				MaybeUnbondCompleteAt: nil,
				Amount:                amount,
				AutoClaimedRewards:    coin.Coin{},
			},
		)}
	}
	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	// When there is no reward withdrew, `transfer` event would not exist
	unbondEvent := log.GetEventByType("unbond")
	if unbondEvent == nil {
		panic("missing `unbond` event in TxsResult log")
	}
	unbondCompletionTime, unbondCompletionTimeErr := utctime.Parse(
		time.RFC3339, unbondEvent.MustGetAttributeByKey("completion_time"),
	)
	if unbondCompletionTimeErr != nil {
		panic(fmt.Sprintf("error parsing unbond completion time: %v", unbondCompletionTimeErr))
	}

	moduleAccounts := tmcosmosutils.NewModuleAccounts(addressPrefix)
	transferEvents := log.GetEventsByType("transfer")
	autoClaimedRewards := coin.NewZeroCoin(stakingDenom)
	for _, transferEvent := range transferEvents {
		sender := transferEvent.MustGetAttributeByKey("sender")
		if sender != moduleAccounts.Distribution {
			continue
		}

		amount := transferEvent.MustGetAttributeByKey("amount")
		coin, coinErr := coin.ParseCoinNormalized(amount)
		if coinErr != nil {
			panic(fmt.Errorf("error parsing auto claimed rewards amount: %v", coinErr))
		}
		autoClaimedRewards = autoClaimedRewards.Add(coin)
	}

	return []command.Command{command_usecase.NewCreateMsgUndelegate(
		msgCommonParams,

		model.MsgUndelegateParams{
			DelegatorAddress:      msg["delegator_address"].(string),
			ValidatorAddress:      msg["validator_address"].(string),
			MaybeUnbondCompleteAt: &unbondCompletionTime,
			Amount:                amount,
			AutoClaimedRewards:    autoClaimedRewards,
		},
	)}
}

func parseMsgBeginRedelegate(
	addressPrefix string,
	stakingDenom string,
	txSuccess bool,
	txsResult model.BlockResultsTxsResult,
	msgIndex int,
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	amountValue, _ := msg["amount"].(map[string]interface{})
	amount := tmcosmosutils.MustNewCoinFromAmountInterface(amountValue)

	if !txSuccess {
		return []command.Command{command_usecase.NewCreateMsgBeginRedelegate(
			msgCommonParams,

			model.MsgBeginRedelegateParams{
				DelegatorAddress:    msg["delegator_address"].(string),
				ValidatorSrcAddress: msg["validator_src_address"].(string),
				ValidatorDstAddress: msg["validator_dst_address"].(string),
				Amount:              amount,
				AutoClaimedRewards:  coin.Coin{},
			},
		)}
	}

	log := utils.NewParsedTxsResultLog(&txsResult.Log[msgIndex])
	moduleAccounts := tmcosmosutils.NewModuleAccounts(addressPrefix)
	transferEvents := log.GetEventsByType("transfer")
	autoClaimedRewards := coin.NewZeroCoin(stakingDenom)
	for _, transferEvent := range transferEvents {
		sender := transferEvent.MustGetAttributeByKey("sender")
		if sender != moduleAccounts.Distribution {
			continue
		}

		amount := transferEvent.MustGetAttributeByKey("amount")
		coin, coinErr := coin.ParseCoinNormalized(amount)
		if coinErr != nil {
			panic(fmt.Errorf("error parsing auto claimed rewards amount: %v", coinErr))
		}
		autoClaimedRewards = autoClaimedRewards.Add(coin)
	}

	return []command.Command{command_usecase.NewCreateMsgBeginRedelegate(
		msgCommonParams,

		model.MsgBeginRedelegateParams{
			DelegatorAddress:    msg["delegator_address"].(string),
			ValidatorSrcAddress: msg["validator_src_address"].(string),
			ValidatorDstAddress: msg["validator_dst_address"].(string),
			Amount:              amount,
			AutoClaimedRewards:  autoClaimedRewards,
		},
	)}
}

func parseMsgUnjail(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgUnjail(
		msgCommonParams,

		model.MsgUnjailParams{
			ValidatorAddr: msg["validator_addr"].(string),
		},
	)}
}

func parseGenesisGenTxsMsgCreateValidator(
	msg map[string]interface{},
) []command.Command {
	amountValue, _ := msg["value"].(map[string]interface{})
	amount := tmcosmosutils.MustNewCoinFromAmountInterface(amountValue)
	tendermintPubkey, _ := msg["pubkey"].(map[string]interface{})
	description := model.ValidatorDescription{
		Moniker:         "",
		Identity:        "",
		Website:         "",
		SecurityContact: "",
		Details:         "",
	}

	if descriptionJSON, ok := msg["description"].(map[string]interface{}); ok {
		description = model.ValidatorDescription{
			Moniker:         descriptionJSON["moniker"].(string),
			Identity:        descriptionJSON["identity"].(string),
			Website:         descriptionJSON["website"].(string),
			SecurityContact: descriptionJSON["security_contact"].(string),
			Details:         descriptionJSON["details"].(string),
		}
	}

	commission := model.ValidatorCommission{
		Rate:          "",
		MaxRate:       "",
		MaxChangeRate: "",
	}
	if commissionJSON, ok := msg["commission"].(map[string]interface{}); ok {
		commission = model.ValidatorCommission{
			Rate:          commissionJSON["rate"].(string),
			MaxRate:       commissionJSON["max_rate"].(string),
			MaxChangeRate: commissionJSON["max_change_rate"].(string),
		}
	}

	return []command.Command{command_usecase.NewCreateGenesisValidator(
		genesis.CreateGenesisValidatorParams{
			// Genesis validator are always bonded
			// TODO: What if gen_txs contains more validators than maximum validators
			Status:            constants.BONDED,
			Description:       description,
			Commission:        commission,
			MinSelfDelegation: msg["min_self_delegation"].(string),
			DelegatorAddress:  msg["delegator_address"].(string),
			ValidatorAddress:  msg["validator_address"].(string),
			TendermintPubkey:  tendermintPubkey["key"].(string),
			Amount:            amount,
			Jailed:            false,
		},
	)}
}

func parseMsgCreateValidator(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	amountValue, _ := msg["value"].(map[string]interface{})
	amount := tmcosmosutils.MustNewCoinFromAmountInterface(amountValue)
	tendermintPubkey, _ := msg["pubkey"].(map[string]interface{})
	description := model.ValidatorDescription{
		Moniker:         "",
		Identity:        "",
		Website:         "",
		SecurityContact: "",
		Details:         "",
	}

	if descriptionJSON, ok := msg["description"].(map[string]interface{}); ok {
		description = model.ValidatorDescription{
			Moniker:         descriptionJSON["moniker"].(string),
			Identity:        descriptionJSON["identity"].(string),
			Website:         descriptionJSON["website"].(string),
			SecurityContact: descriptionJSON["security_contact"].(string),
			Details:         descriptionJSON["details"].(string),
		}
	}

	commission := model.ValidatorCommission{
		Rate:          "",
		MaxRate:       "",
		MaxChangeRate: "",
	}
	if commissionJSON, ok := msg["commission"].(map[string]interface{}); ok {
		commission = model.ValidatorCommission{
			Rate:          commissionJSON["rate"].(string),
			MaxRate:       commissionJSON["max_rate"].(string),
			MaxChangeRate: commissionJSON["max_change_rate"].(string),
		}
	}

	return []command.Command{command_usecase.NewCreateMsgCreateValidator(
		msgCommonParams,

		model.MsgCreateValidatorParams{
			Description:       description,
			Commission:        commission,
			MinSelfDelegation: msg["min_self_delegation"].(string),
			DelegatorAddress:  msg["delegator_address"].(string),
			ValidatorAddress:  msg["validator_address"].(string),
			TendermintPubkey:  tendermintPubkey["key"].(string),
			Amount:            amount,
		},
	)}
}

func parseMsgEditValidator(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	var description model.ValidatorDescription
	if descriptionJSON, ok := msg["description"].(map[string]interface{}); ok {
		description = model.ValidatorDescription{
			Moniker:         descriptionJSON["moniker"].(string),
			Identity:        descriptionJSON["identity"].(string),
			Website:         descriptionJSON["website"].(string),
			SecurityContact: descriptionJSON["security_contact"].(string),
			Details:         descriptionJSON["details"].(string),
		}
	}

	var maybeCommissionRate *string
	if msg["commission_rate"] != nil {
		maybeCommissionRate = primptr.String(msg["commission_rate"].(string))
	}

	var maybeMinSelfDelegation *string
	if msg["min_self_delegation"] != nil {
		maybeMinSelfDelegation = primptr.String(msg["min_self_delegation"].(string))
	}

	return []command.Command{command_usecase.NewCreateMsgEditValidator(
		msgCommonParams,

		model.MsgEditValidatorParams{
			Description:            description,
			ValidatorAddress:       msg["validator_address"].(string),
			MaybeCommissionRate:    maybeCommissionRate,
			MaybeMinSelfDelegation: maybeMinSelfDelegation,
		},
	)}
}

func parseMsgNFTIssueDenom(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgNFTIssueDenom(
		msgCommonParams,

		model.MsgNFTIssueDenomParams{
			DenomId:   msg["id"].(string),
			DenomName: msg["name"].(string),
			Schema:    msg["schema"].(string),
			Sender:    msg["sender"].(string),
		},
	)}
}

func parseMsgNFTMintNFT(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgNFTMintNFT(
		msgCommonParams,

		model.MsgNFTMintNFTParams{
			DenomId:   msg["denom_id"].(string),
			TokenId:   msg["id"].(string),
			TokenName: msg["name"].(string),
			URI:       msg["uri"].(string),
			Data:      msg["data"].(string),
			Sender:    msg["sender"].(string),
			Recipient: msg["recipient"].(string),
		},
	)}
}

func parseMsgNFTTransferNFT(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgNFTTransferNFT(
		msgCommonParams,

		model.MsgNFTTransferNFTParams{
			TokenId:   msg["id"].(string),
			DenomId:   msg["denom_id"].(string),
			Sender:    msg["sender"].(string),
			Recipient: msg["recipient"].(string),
		},
	)}
}

func parseMsgNFTEditNFT(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgNFTEditNFT(
		msgCommonParams,

		model.MsgNFTEditNFTParams{
			DenomId:   msg["denom_id"].(string),
			TokenId:   msg["id"].(string),
			TokenName: msg["name"].(string),
			URI:       msg["uri"].(string),
			Data:      msg["data"].(string),
			Sender:    msg["sender"].(string),
		},
	)}
}

func parseMsgNFTBurnNFT(
	msgCommonParams event.MsgCommonParams,
	msg map[string]interface{},
) []command.Command {
	return []command.Command{command_usecase.NewCreateMsgNFTBurnNFT(
		msgCommonParams,

		model.MsgNFTBurnNFTParams{
			DenomId: msg["denom_id"].(string),
			TokenId: msg["id"].(string),
			Sender:  msg["sender"].(string),
		},
	)}
}
