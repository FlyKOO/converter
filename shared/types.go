package shared

// 定义交易解析过程中使用的结构体

import pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"

// TransactionDetails 描述一笔交易的详细信息
type TransactionDetails struct {
	Signature            string               // 交易签名
	ComputeUnitsConsumed uint64               // 消耗的计算单位
	Slot                 uint64               // 所在区块槽
	Instructions         []InstructionDetails // 指令列表
	BalanceChanges       []BalanceChanges     // SOL 余额变化
	TokenBalanceChanges  TokenBalanceChanges  // 代币余额变化
	Logs                 []string             // 执行日志
}

// TokenBalanceChanges 记录交易前后的代币余额
type TokenBalanceChanges struct {
	PreTokenBalances  []*pb.TokenBalance // 交易前的代币余额
	PostTokenBalances []*pb.TokenBalance // 交易后的代币余额
}

// BalanceChanges 记录账户 SOL 余额的变化
type BalanceChanges struct {
	Account       AccountReference // 账户信息
	BalanceBefore uint64           // 交易前余额
	BalanceAfter  uint64           // 交易后余额
}

// InstructionDetails 描述单条指令的执行情况
type InstructionDetails struct {
	Index             int                       // 指令序号
	ProgramID         AccountReference          // 程序账户
	Data              []byte                    // 指令数据
	Accounts          []AccountReference        // 关联账户
	InnerInstructions []InnerInstructionDetails // 内部指令列表
}

// InnerInstructionDetails 描述内层指令的执行情况
type InnerInstructionDetails struct {
	OuterIndex int                // 所属外层指令序号
	InnerIndex int                // 内部指令序号
	ProgramID  AccountReference   // 程序账户
	Data       []byte             // 指令数据
	Accounts   []AccountReference // 关联账户
}

// AccountReference 描述账户在交易中的属性
type AccountReference struct {
	PublicKey  string // 账户公钥
	IsWritable bool   // 是否可写
	IsReadable bool   // 是否只读
	IsSigner   bool   // 是否为签名者
	IsLut      bool   // 是否来自地址查找表
}
