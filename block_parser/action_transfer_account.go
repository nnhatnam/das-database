package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/scorpiotzh/toolib"
	"strconv"
)

func (b *BlockParser) ActionTransferAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version transfer account tx")
		return
	}

	log.Info("ActionTransferAccount:", req.BlockNumber, req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := builder.Account

	oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)
	oldBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	res, err := b.ckbClient.GetTxByHashOnChain(req.Tx.Inputs[oldBuilder.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTxByHashOnChain err: %s", err.Error())
		return
	}

	_, _, oldChainType, _, oldAddress, _ := core.FormatDasLockToHexAddress(res.Transaction.Outputs[oldBuilder.Index].Lock.Args)
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        account,
		Action:         common.DasActionTransferAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oldChainType,
		Address:        oldAddress,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(builder.Index)),
		Account:            account,
		OwnerChainType:     oCT,
		Owner:              oA,
		OwnerAlgorithmId:   oID,
		ManagerChainType:   mCT,
		Manager:            mA,
		ManagerAlgorithmId: mID,
	}
	var recordsInfos []dao.TableRecordsInfo
	recordList := builder.RecordList()
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			Account: account,
			Key:     v.Key,
			Type:    v.Type,
			Label:   v.Label,
			Value:   v.Value,
			Ttl:     strconv.FormatUint(uint64(v.TTL), 10),
		})
	}

	log.Info("ActionTransferAccount:", account, oID, mID, oCT, mCT, oA, mA, transactionInfo.Address)

	if err := b.dbDao.TransferAccount(accountInfo, transactionInfo, recordsInfos); err != nil {
		log.Error("TransferAccount err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("TransferAccount err: %s", err.Error())
	}

	return
}
