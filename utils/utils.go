package utils

// 与交易解析相关的辅助函数

import (
	"bytes"
	"errors"

	"github.com/FlyKOO/converter/shared"
	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

// ProcessTransactionToStruct 将订阅的交易更新转换为结构体信息
func ProcessTransactionToStruct(tx *pb.SubscribeUpdateTransaction, signature string) (*shared.TransactionDetails, error) {
	if tx == nil || tx.Transaction == nil {
		return nil, errors.New("交易为空")
	}

	meta := tx.Transaction.GetMeta()
	msg := tx.GetTransaction().GetTransaction().GetMessage()
	if msg == nil {
		return nil, errors.New("交易消息为空")
	}

	// 创建交易详情结构体，填充基本信息
	txDetails := &shared.TransactionDetails{
		Signature:            signature,
		ComputeUnitsConsumed: meta.GetComputeUnitsConsumed(),
		Slot:                 tx.Slot,
		Logs:                 meta.GetLogMessages(),
		TokenBalanceChanges: shared.TokenBalanceChanges{
			PreTokenBalances:  meta.PreTokenBalances,
			PostTokenBalances: meta.PostTokenBalances,
		},
	}

	// 处理账户列表并构建快速查询映射
	accountKeys := msg.GetAccountKeys()
	signersIndexes := msg.Header.GetNumRequiredSignatures()
	readSigners := msg.Header.GetNumReadonlySignedAccounts()
	readNonSigners := msg.Header.GetNumReadonlyUnsignedAccounts()

	// 创建账户属性映射，便于快速访问
	accountMap := make(map[string]shared.AccountReference, len(accountKeys))
	accountList := make([]shared.AccountReference, len(accountKeys))

	for i, account := range accountKeys {
		pubKey := base58.Encode(account)
		ref := shared.AccountReference{
			PublicKey: pubKey,
		}

		switch {
		case i < int(signersIndexes):
			// 可写签名账户
			ref.IsWritable = true
			ref.IsSigner = true
		case i < int(signersIndexes+readSigners):
			// 只读签名账户
			ref.IsReadable = true
			ref.IsSigner = true
		case i < int(int(signersIndexes)+int(readSigners)+(len(accountKeys)-int(signersIndexes+readSigners+readNonSigners))):
			// 可写非签名账户
			ref.IsWritable = true
		default:
			// 只读非签名账户
			ref.IsReadable = true
		}

		accountList[i] = ref
		accountMap[pubKey] = ref
	}

	// 处理余额变化
	preBalances := meta.GetPreBalances()
	postBalances := meta.GetPostBalances()
	balanceChanges := make([]shared.BalanceChanges, len(accountKeys))

	loadedWritable := meta.GetLoadedWritableAddresses()
	loadedReadonly := meta.GetLoadedReadonlyAddresses()

	for i, account := range accountKeys {
		pubKey := base58.Encode(account)
		accRef := accountMap[pubKey]

		// 根据加载的地址更新可读写状态
		if contains(loadedWritable, account) {
			accRef.IsWritable = true
			accRef.IsLut = true
		}
		if contains(loadedReadonly, account) {
			accRef.IsReadable = true
			accRef.IsLut = true
		}

		balanceChanges[i] = shared.BalanceChanges{
			Account:       accRef,
			BalanceBefore: preBalances[i],
			BalanceAfter:  postBalances[i],
		}
	}
	txDetails.BalanceChanges = balanceChanges

	// 为指令解析创建合并后的账户列表
	// 预分配所需容量
	mergedAccounts := make([][]byte, 0, len(accountKeys)+
		len(loadedWritable)+
		len(loadedReadonly))
	mergedAccounts = append(mergedAccounts, accountKeys...)
	mergedAccounts = append(mergedAccounts, loadedWritable...)
	mergedAccounts = append(mergedAccounts, loadedReadonly...)

	// 解析指令
	instructions := msg.GetInstructions()
	txDetails.Instructions = make([]shared.InstructionDetails, len(instructions))

	for idx, inst := range instructions {
		programIdIndex := inst.GetProgramIdIndex()
		if int(programIdIndex) >= len(mergedAccounts) {
			// zap.L().Fatal("程序 ID 索引无效",
			// 	zap.Uint32("program_id_index", programIdIndex),
			// 	zap.Int("accounts_length", len(mergedAccounts)))
			continue
		}

		programID := mergedAccounts[programIdIndex]
		programIDStr := base58.Encode(programID)

		instruction := shared.InstructionDetails{
			Index: idx + 1,
			ProgramID: shared.AccountReference{
				PublicKey:  programIDStr,
				IsWritable: contains(loadedWritable, programID) || accountMap[programIDStr].IsWritable,
				IsReadable: contains(loadedReadonly, programID) || accountMap[programIDStr].IsReadable,
				IsSigner:   accountMap[programIDStr].IsSigner,
				IsLut:      contains(loadedWritable, programID) || contains(loadedReadonly, programID),
			},
			Data: inst.GetData(),
		}

		// 处理指令中引用的账户
		accounts := inst.GetAccounts()
		instruction.Accounts = make([]shared.AccountReference, len(accounts))

		for accIdx, accIndex := range accounts {
			if int(accIndex) >= len(mergedAccounts) {
				// zap.L().Fatal("账户索引无效",
				// zap.Int("account_index", int(accIndex)),
				// zap.Int("accounts_length", len(mergedAccounts)))
				continue
			}

			accBytes := mergedAccounts[accIndex]
			accStr := base58.Encode(accBytes)

			instruction.Accounts[accIdx] = shared.AccountReference{
				PublicKey:  accStr,
				IsWritable: contains(loadedWritable, accBytes) || accountMap[accStr].IsWritable,
				IsReadable: contains(loadedReadonly, accBytes) || accountMap[accStr].IsReadable,
				IsSigner:   accountMap[accStr].IsSigner,
				IsLut:      contains(loadedWritable, accBytes) || contains(loadedReadonly, accBytes),
			}
		}

		// 处理内层指令
		if innerInsts := getInnerInstructions(meta, uint32(idx)); len(innerInsts) > 0 {
			instruction.InnerInstructions = make([]shared.InnerInstructionDetails, len(innerInsts))

			for innerIdx, inner := range innerInsts {
				innerProgramIdIndex := inner.GetProgramIdIndex()
				if int(innerProgramIdIndex) >= len(mergedAccounts) {
					// zap.L().Fatal("内层程序 ID 索引无效",
					// 	zap.Uint32("program_id_index", innerProgramIdIndex),
					// 	zap.Int("accounts_length", len(mergedAccounts)))
					continue
				}

				innerInstruction := shared.InnerInstructionDetails{
					OuterIndex: idx + 1,
					InnerIndex: innerIdx + 1,
					ProgramID: shared.AccountReference{
						PublicKey:  base58.Encode(mergedAccounts[innerProgramIdIndex]),
						IsWritable: contains(loadedWritable, programID) || accountMap[programIDStr].IsWritable,
						IsReadable: contains(loadedReadonly, programID) || accountMap[programIDStr].IsReadable,
						IsSigner:   accountMap[programIDStr].IsSigner,
					},
					Data: inner.GetData(),
				}

				// 处理内层指令涉及的账户
				innerAccounts := inner.GetAccounts()
				innerInstruction.Accounts = make([]shared.AccountReference, len(innerAccounts))

				for accIdx, accIndex := range innerAccounts {
					if int(accIndex) >= len(mergedAccounts) {
						// zap.L().Fatal("内层账户索引无效",
						// 	zap.Int("account_index", int(accIndex)),
						// 	zap.Int("accounts_length", len(mergedAccounts)))
						continue
					}

					accBytes := mergedAccounts[accIndex]
					accStr := base58.Encode(accBytes)

					innerInstruction.Accounts[accIdx] = shared.AccountReference{
						PublicKey:  accStr,
						IsWritable: contains(loadedWritable, accBytes) || accountMap[accStr].IsWritable,
						IsReadable: contains(loadedReadonly, accBytes) || accountMap[accStr].IsReadable,
						IsSigner:   accountMap[accStr].IsSigner,
					}
				}

				instruction.InnerInstructions[innerIdx] = innerInstruction
			}
		}

		txDetails.Instructions[idx] = instruction
	}

	return txDetails, nil
}

// contains 判断字节切片 s 中是否包含元素 e
func contains(s [][]byte, e []byte) bool {
	for _, a := range s {
		if bytes.Equal(a, e) {
			return true
		}
	}
	return false
}

// getInnerInstructions 根据索引获取内层指令列表
func getInnerInstructions(meta *pb.TransactionStatusMeta, index uint32) []*pb.InnerInstruction {
	if meta == nil {
		return nil
	}
	for _, inner := range meta.GetInnerInstructions() {
		if inner.GetIndex() == index {
			return inner.GetInstructions()
		}
	}
	return nil
}
