<!--
order: 2
-->

# State

## DoPoolDelta

Market module provides swap functionality based on constant product mechanism. do pool have to keep its delta to track the currency demands for swap spread. Do pool can be retrived from do pool delta with following equation:

```go
DoPool := BasePool + delta
LunaPool := (BasePool * BasePool) / DoPool
```

> Note that the all pool holds decimal unit of `usdr` amount, so delta is also `usdr` unit.

- DoPoolDelta: `0x01 -> amino(DoPoolDelta)`

```go
type DoPoolDelta sdk.Dec // the gap between the DoPool and the BasePool
```






