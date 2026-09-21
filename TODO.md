# QuickValidate TODO List

对齐目标：go-playground/validator **v10.25.0** 的内置 tag 行为，以 `validator.New()`（默认配置）为准。

## 已完成

### 引擎

- [x] 解析器重写：`validate` tag 严格按 validator 的 `parseFieldTagsRecursive` 解析（别名、`|` 或分支、`dive`/`keys`/`endkeys`、`0x2C`/`0x7C` 转义）
- [x] 静态类型推断：字段 kind、元素类型、`time.Time`/`time.Duration`、同包命名类型的底层类型
- [x] 生成器改为 tag 注册表 + 逐 tag 发射器（原先是一个巨型模板里的 if-else 链）
- [x] 遍历语义与 validator 的 `traverseField` 对齐：每个字段只报第一个失败的 tag、`omitempty`/`omitnil`/`omitzero` 门控、nil 指针处理（含 `runValidationOnNil` 白名单）、`dive`/`keys`/`endkeys`、嵌套结构体自动递归
- [x] 命名空间与 Field 名带运行期真实下标（`User.Items[3]`、`User.M[key]`）

### Tag 覆盖

- [x] 比较与长度：`len`/`min`/`max`/`eq`/`ne`/`gt`/`gte`/`lt`/`lte`/`eq_ignore_case`/`ne_ignore_case`（字符串按字符数、字节长度差异、`time.Duration` 参数、`time.Time` 与当前时间比较）
- [x] 容器：`dive`、`keys`、`endkeys`、`unique`（切片/数组/map/按字段去重）
- [x] 跨字段：`eqfield`/`nefield`/`gtfield`/`gtefield`/`ltfield`/`ltefield` 及 `cs` 变体、`fieldcontains`/`fieldexcludes`
- [x] 条件必填：`required_if`/`required_unless`/`skip_unless`/`required_with*`/`required_without*`、`excluded_*`
- [x] 字符串内容：`contains*`/`excludes*`/`startswith`/`endswith`/`startsnotwith`/`endsnotwith`
- [x] 全部内置 tag：对照 v10.25.0 源码逐个核对，`bakedInValidators`(168) + 别名(3) + 结构 tag(8) = 179 个，生成器 `knownTags()` 与之双向差集为空
- [x] 读数值的 tag（`iso3166_1_alpha_numeric*`、`iso4217_numeric`、`port`）按 `field.Kind()` 分派到字符串/有符号/无符号入口，与上游分支逐条对应（无符号国家码先 `%1000` 再转型，币种码直接转型）
- [x] 别名 `iscolor`/`country_code`/`eu_country_code`
- [x] `oneof`/`oneofci`（含整数/无符号字段的十进制字符串比较）

### 运行时辅助包

- [x] `pkg/validators`：从 validator 逐字移植的无反射实现（4 个并行子任务完成，各自用真实库做 oracle 自测：a1 3667 例、a2 145419 例、a3 468 例、a4 1166 例）
- [x] 正则表与数据表（`regexes.go`、`country_codes.go`、`currency_codes.go`、`postcode_regexes.go`）原样移植并注明出处

### 验证

- [x] `internal/align` 差分测试套件：手写边界用例 + 随机差分（按类型取值池填充结构体），逐条比对 `命名空间|tag|参数`
- [x] 文件系统类 tag（`file`/`dir`/`filepath`/`dirpath`/`image`）用真实临时文件验证
- [x] 示例（simple/numeric/benchmark）改为使用生成代码
- [x] 语料覆盖 179/179 个内置 tag；单次运行 5021 条比对，跳过 60 条（全部是上游 panic 的 `unique` + nil 指针一类）
- [x] `pkg/generator` 单元测试：kind 分派结果、不支持的 kind 报错、spicedb 参数校验、重复注册 panic

### 本轮修复（均由差分套件暴露）

- [x] `iso4217_numeric` 的专属发射器被 `registry.go` 的批量表按 init 文件序**静默覆盖**，整型字段生成不可编译代码；改为 `register()` 重复注册即 panic，并移除批量表里的重复项
- [x] `iso3166_1_alpha_numeric*` `iso4217_numeric` 在整型字段上生成不可编译代码：补 `Int`/`Uint` 入口，`kindHelper` 按 kind 发射显式转换
- [x] 同包命名类型（`type Email string`）传参给只收内建类型的 helper 时不可编译：`TypeInfo.Named` + `fieldRef.arg()` 在函数实参位置做显式转换（比较、`len()`、方法调用保持原样）
- [x] 语料里 `spicedb=user` 参数非法，让上游 panic 并连坐跳过整个 `Formats2` 结构体（303 例）；改用合法参数，并在生成期拒绝非法参数
- [x] **嵌入字段（匿名）此前被整个丢掉**：`type Outer struct { Base; ... }` 的生成代码里没有 `s.Base`，`Base` 上的校验全部静默失效（validator 会报 `Outer.Base.Common`）。解析器不再跳过匿名字段，按 `reflect.StructField.Name` 用类型名做字段名（`embeddedFieldName`），未导出字段的跳过条件也改成与 validator 一致（`!Anonymous && unexported`）
- [x] **`interface{}` 字段此前完全不生成**：`emitNullable` 因 interface 没有 `Elem` 直接返回。现在按"只依赖是否为 nil"的 tag 子集生成（`required`/`isdefault`/`required_*`/`excluded_*`/`skip_unless`/`omit*`），其余 tag 在生成期报错，不再静默漏过
- [x] **字符串类 tag 的生成期类型检查补全**：`email`/`uuid`/`datetime`/`spicedb`/`postcode_*` 等只读字符串的 tag 用在非字符串字段上，此前会生成不可编译代码（`qv.IsEmail(s.Int)`）；现在统一报 `xxx requires a string field, got int`。validator 在这些组合下不会 panic，而是通过反射的 `String()` 读到 `"<int Value>"` 之类的占位串，于是该字段永远报失败——报错比这更有用
- [x] **`json` 支持 `[]byte`**：参考实现的 `isJSON` 有 Slice 分支，qv 此前只当字符串处理，`Doc []byte validate:"json"` 生成不可编译代码；现在按 kind 分派到新增的 `IsJSONBytes`，命名 `[]byte` 类型做 `[]byte(...)` 转换
- [x] **错误对象补齐 `ActualTag` / `StructNamespace` / `Kind`**：`ActualTag` 是别名的展开（`iscolor` → `hexcolor|rgb|rgba|hsl|hsla`）或用户写的整段或分支；`StructNamespace` 与 `Namespace` 在无 `TagNameFunc` 时恒等；`Kind` 是解引用后的 reflect kind（nil 指针报 `ptr`，接口字段取运行期动态类型）。差分套件已扩展为逐条比对这六个字段（5021 条），顺带查出 `int`/`int64`、`uint`/`uint64` 在 `BuiltinName` 里被混为一谈的问题
- [x] **跨字段路径经过指针**：`eqfield=A.B` 且 `A` 为 `*T` 此前在生成期被整个拒绝（一律判失败）。现在按参考实现的语义生成——路径上有指针时加非 nil 守卫，找不到（指针为 nil、字段不存在、kind 不匹配）时**否定型比较判通过、肯定型判失败**（`nefield`/`fieldexcludes` 与 `eqfield`/`fieldcontains` 相反）
- [x] `pkg/parser` 单元测试（tag 链、别名展开、或分支、`0x2C`/`0x7C` 转义、`dive`/`keys`、`BuiltinName`、嵌入字段收集）与 `examples/benchmark` 的 `testing.B` 基准（扁平 480ns/0 分配 vs 参考 631ns/1 分配；100 地址嵌套 2384ns/0 分配 vs 10554ns/101 分配）
- [x] 跨字段比较两侧类型不同时（具名 vs 内建、宽度不同）生成不可编译代码。按参考实现实际的 reflect 读取语义分派：`eqfield`/`nefield`/`gtfield` 系列按底层值比较（两侧一起转型到共同类型），`unique=Other` 按 `interface{}` 比较（类型不同恒不相等，直接判 true），reflect kind 不同（如 `int8` 对 `int64`，或 `int` 对 `int64`）直接判 false。`time.Duration` 记为 `int64`，与 `reflect.Int64` 一致

## 已知限制

- [ ] 自定义校验函数（`RegisterValidation`）、结构体级校验（`RegisterStructValidation`）、翻译器不支持
- [ ] 跨包结构体的嵌套递归：只能递归同包内由本工具生成的结构体
- [ ] `unique` 作用于含 nil 指针元素的切片会让上游库 panic，差分测试标记跳过（当前 60 例，是唯一剩下的跳过类别）
- [ ] `timezone`/`bcp47_language_tag`/`*_addr` 的结果依赖运行时环境，与 validator 行为一致但不可完全确定
- [ ] `interface{}` 字段上依赖动态类型的 tag（`min`/`email`/`dive` 等）在生成期报错，而 validator 会在运行期按接口里的实际值校验。这是静态生成固有的取舍（与"运行期不反射"冲突），目前选择"报错"而不是"静默漏过"；如需支持，只能让生成代码在该字段上回退到反射

## 后续可做

- [ ] 支持 `WithRequiredStructEnabled()` 等选项，让生成代码可切换配置
