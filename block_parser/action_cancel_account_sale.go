package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
)

func (b *BlockParser) ActionCancelAccountSale(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version cancel account sale tx")
		return
	}

	log.Info("ActionCancelAccountSale:", req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:   builder.AccountId,
		Account:     builder.Account,
		Status:      dao.AccountStatusNormal,
	}

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionCancelAccountSale,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.Outputs[1].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 1),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionCancelAccountSale:", transactionInfo.Account)

	if err := b.dbDao.CancelAccountSale(accountInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("CancelAccountSale err: %s", err.Error())
		return
	}

	return
}
