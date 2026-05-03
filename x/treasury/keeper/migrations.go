package keeper

import (
	"github.com/Daviddochain/dochain-core/v4/x/treasury/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var burnTaxExcemptionAddressList = []string{
	"do10atxpzafqfjy58z0dvugmd9zf63fycr6uvwhjm",
	"do1jrq7xa63a4qgpdgtj70k8yz5p32ps9r7mlj3yr",
	"do15s66unmdcpknuxxldd7fsr44skme966tdckq8c",
	"do1u0p7xuwlg0zsqgntagdkyjyumsegd8agzhug99",
	"do1fax8l6srhew5tu2mavmu83js3v7vsqf9yr4fv7",
	"do132wegs0kf9q65t9gsm3g2y06l98l2k4treepkq",
	"do1l89hprfccqxgzzypjzy3fnp7vqpnkqg5vvqgjc",
	"do1ns7lfvrxzter4d2yl9tschdwntcxa25vtsvd8a",
	"do1vuvju6la7pj6t8d8zsx4g8ea85k2cg5u62cdhl",
	"do1lzdux37s4anmakvg7pahzh03zlf43uveq83wh2",
	"do1ky3qcf7v45n6hwfmkm05acwczvlq8ahnq778wf",
	"do17m8tkde0mav43ckeehp537rsz4usqx5jayhf08",
	"do1urj8va62jeygra7y3a03xeex49mjddh3eul0qa",
	"do10wyptw59xc52l86pg86sy0xcm3nm5wg6a3cf7l",
	"do1sujaqwaw7ls9fh6a4x7n06nv7fxx5xexwlnrkf",
	"do1qg59nhvag222kp6fyzxt83l4sw02huymqnklww",
	"do1dxxnwxlpjjkl959v5xrghx0dtvut60eef6vcch",
	"do1y246m036et7vu69nsg4kapelj0tywe8vsmp34d",
	"do1j39c9sjr0zpjnrfjtthuua0euecv7txavxvq36",
	"do1t0jthtq9zhm4ldtvs9epp02zp23f355wu6zrzq",
	"do12dxclvqrgt7w3s7gtwpdkxgymexv8stgqcr0yu",
	"do1az3dsad74pwhylrrexnn5qylzj783uyww2s7xz",
	"do1ttq26dq4egr5exmhd6gezerrxhlutx9u90uncn",
	"do13e9670yuvfs06hctt9pmgjnz0yw28p0wgnhrqn",
	"do1skmktm537pfaycgu9jx4fqryjt6pf77ycpesw0",
	"do14q8cazgt58y2xkd26mlukemwth0cnvfqmgz2qk",
	"do163vzxz9wwy320ccwy73qe6h33yzg2yhyvv5nsf",
	"do1kj43wfnvrgc2ep94dgmwvnzv8vnkkxrxmrnhkp",
	"do1gu6re549pn0mdpshtv75t3xugn347jghlhul73",
	"do1gft3qujlq04yza3s2r238mql2yn3xxqepzt2up",
	"do174pe7qe7g867spzdfs5f4rf9fuwmm42zf4hykf",
	"do1ju68sg6k39t385sa0fazqvjgh6m6gkmsmp4lln",
	"do1dlh7k4hcnsrvlfuzhdzx3ctynj7s8dde9zmdyd",
	"do18wcdhpzpteharlkks5n6k7ent0mjyftvcpm6ee",
	"do1xmkwsauuk3kafua9k23hrkfr76gxmwdfq5c09d",
	"do1t957gces65xd6p8g4cuqnyd0sy5tzku59njydd",
	"do1s4rd0y5e4gasf0krdm2w8sjhsmh030m74f2x9v",
	"do15jya6ugxp65y80y5h82k4gv90pd7acv58xp6jj",
	"do14yqy9warjkxyecda5kf5a68qlknf4ve4sh7sa6",
	"do1yxras4z0fs9ugsg2hew9334k65uzejwcslyx0y",
	"do1p0vl4s4gp46vy6dm352s2fgtw6hccypph7zc3u",
	"do1hhj92twle9x8rjkr3yffujexsy5ldexak5rglz",
	"do18vnrzlzm2c4xfsx382pj2xndqtt00rvhu24sqe",
	"do1ncjg4a59x2pgvqy9qjyqprlj8lrwshm0wleht5",
	"do19l7hzwazq5j0dykfldcwrk2927xwcjd0kt0vt9",
	"do1frh79vmtur5fmrghz6gfjvfhpa3u2c0uemv4af",
}

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	m.keeper.SetBurnSplitRate(ctx, types.DefaultBurnTaxSplit)

	for _, address := range burnTaxExcemptionAddressList {
		m.keeper.AddBurnTaxExemptionAddress(ctx, address)
	}

	return nil
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	m.keeper.SetMinInitialDepositRatio(ctx, types.DefaultMinInitialDepositRatio)

	return nil
}






