<!--
order: 2
-->

# State

## TerraPoolDelta

Market module provides swap functionality based on constant product mechanism. dochain pool have to keep its delta to track the currency demands for swap spread. Luna pool can be retrived from dochain pool delta with following equation:

```go
TerraPool := BasePool + delta
LunaPool := (BasePool * BasePool) / TerraPool
```

> Note that the all pool holds decimal unit of `usdr` amount, so delta is also `usdr` unit.

- TerraPoolDelta: `0x01 -> amino(TerraPoolDelta)`

```go
type TerraPoolDelta sdk.Dec // the gap between the TerraPool and the BasePool
```




