package detector

type SelectorInfo struct {
	Label    string
	MethodID int
}

var KnownSelectors = map[string]SelectorInfo{
	"0x022c0d9f": {Label: "UniV2.swap(uint256,uint256,address,bytes)", MethodID: 2},
	"0x128acb08": {Label: "UniV3.swap(address,bool,int256,uint160,bytes)", MethodID: 3},
	"0xa83064c5": {Label: "FluidDex.swapIn(bool,uint256,uint256,address)", MethodID: 4},
	"0x48c89491": {Label: "V4PoolManager.unlock(bytes)", MethodID: 7},
	"0xf39b5b9b": {Label: "UniV1.ethToTokenSwapInput", MethodID: 8},
	"0xad65d76d": {Label: "UniV1.tokenToEthTransferInput", MethodID: 8},
	"0x95e3c50b": {Label: "UniV1.ethToTokenTransferInput", MethodID: 8},
	"0xc52f3057": {Label: "MaverickV2.swap", MethodID: 9},
	"0x3df02124": {Label: "Curve.exchange(int128,int128,uint256,uint256)", MethodID: 10},
	"0xa6417ed6": {Label: "Curve.exchange_underlying(int128,int128,uint256,uint256)", MethodID: 10},
	"0xddc1f59d": {Label: "Curve.exchange(uint256,uint256,uint256,uint256)", MethodID: 11},
	"0x65b2489b": {Label: "Curve.exchange_underlying(uint256,uint256,uint256,uint256)", MethodID: 11},
	"0xc1a3f5e2": {Label: "AllBridge.swap", MethodID: 13},
	"0x95991276": {Label: "MakerDao.sellGem", MethodID: 14},
	"0x8d7ef9bb": {Label: "MakerDao.buyGem", MethodID: 14},
	"0xf95cb83b": {Label: "MakerDao.usdsToDai", MethodID: 14},
	"0x2c1f5cba": {Label: "MakerDao.daiToUsds", MethodID: 14},
	"0x6c197ff5": {Label: "MetronomeSynth.swap(address,address,uint256)", MethodID: 15},
	"0xe43581b8": {Label: "BancorVault.tradeBySourceAmount", MethodID: 17},
	"0xb77d239b": {Label: "BancorConverter.convertByPath", MethodID: 18},
	"0xbd6015b4": {Label: "DODO.sellBase", MethodID: 21},
	"0xdd93f59a": {Label: "DODO.sellQuote", MethodID: 21},
	"0x52bbbe29": {Label: "BalancerVault.swap", MethodID: 30},
	"0x945bcec9": {Label: "BalancerVault.batchSwap", MethodID: 30},
	"0x5c38449e": {Label: "BalancerVault.flashLoan", MethodID: 30},
	"0xc4805a1f": {Label: "1inchV6.fillOrderArgs", MethodID: 44},
	"0x3f62192e": {Label: "UniswapXReactor.execute", MethodID: 45},
	"0x00a718a9": {Label: "AaveV3.liquidationCall", MethodID: 50},
	"0xb858183f": {Label: "SolidlyV3.swap(address,bool,int256,uint160)", MethodID: 71},
	"0xd0e30db0": {Label: "WETH.deposit", MethodID: 250},
	"0x2e1a7d4d": {Label: "WETH.withdraw", MethodID: 250},
}

var TokenAndUtilitySelectors = map[string]string{
	"0xa9059cbb": "ERC20.transfer",
	"0x23b872dd": "ERC20.transferFrom",
	"0x095ea7b3": "ERC20.approve",
	"0x70a08231": "ERC20.balanceOf",
	"0x06fdde03": "ERC20.name",
	"0x95d89b41": "ERC20.symbol",
	"0x313ce567": "ERC20.decimals",
	"0x18160ddd": "ERC20.totalSupply",
	"0xdd62ed3e": "ERC20.allowance",
	"0x0dfe1681": "Pool.token0",
	"0xd21220a7": "Pool.token1",
	"0x0902f1ac": "V2Pair.getReserves",
	"0x3850c7bd": "V3Pool.slot0",
	"0xddca3f43": "V3Pool.fee",
	"0x1a686502": "V3Pool.liquidity",
	"0xd0c93a7c": "V3Pool.tickSpacing",
	"0xfa461e33": "uniswapV3SwapCallback",
	"0x10d1e85c": "uniswapV2Call",
	"0x23e30c8b": "onFlashLoan",
	"0xf04f2707": "receiveFlashLoan",
	"0xe9cbafb0": "uniswapV3FlashCallback",
	"0xdebe51d3": "maverickV2SwapCallback",
}

var SupportedMethodIDs = map[int]bool{
	2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true,
	9: true, 10: true, 11: true, 13: true, 14: true, 15: true,
	17: true, 18: true, 21: true, 22: true, 30: true, 44: true,
	71: true, 250: true,
}

var RemovedOrUnsupportedMethodIDs = map[int]bool{
	45: true,
	50: true,
}
