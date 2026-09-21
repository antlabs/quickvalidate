# QuickValidate

[English](README.md) | [中文](README_zh.md)

QuickValidate 是 [go-playground/validator](https://github.com/go-playground/validator) 的**静态代码生成版**：读取结构体上的 `validate` tag，在构建期生成校验代码，校验一个结构体不需要用反射去遍历它。

生成代码的 **tag 行为与 go-playground/validator v10.25.0 完全对齐**（同一组取值下报出的失败字段、tag、参数、命名空间一致），并由仓库内的差分测试持续验证。

## 特性

- 与 go-playground/validator **同一套 tag 语法与语义**：`required`、`omitempty`、`min/max`、`dive`、`keys/endkeys`、跨字段、条件必填，以及 v10.25.0 的**全部 179 个内置 tag**（`bakedInValidators` + 别名 + 结构 tag，对照上游源码逐个核对）
- 生成期完成参数解析与类型分派，运行期只有直接的比较与函数调用
- 支持同包命名类型（`type Email string`）、指针、切片、map 元素，以及**嵌入结构体**（命名空间与 validator 一致：`Outer.Base.Field`，嵌入指针同样处理）
- 类型与参数错误在**生成期**暴露（validator 在运行期 panic）
- 错误对象与参考库一致：`Namespace`、`StructNamespace`、`Field`、`Tag`、`ActualTag`（别名报展开后的写法）、`Param`、`Kind`、`Value`
- **校验路径不依赖反射**：生成代码只做比较与函数调用，运行时辅助包本身无反射。用到 `reflect` 的只有三处——错误里的 `Kind` 常量、结构体/数组字段的存在性判断（参考库也是这么做的）、以及接口字段的错误元信息
- 自带**差分测试**：同一批结构体与取值，逐条比对生成代码与真实库的输出

## 安装

```bash
go install github.com/antlabs/quickvalidate/cmd/quickvalidate@latest
```

## 使用方法

1. 照常写结构体（与 go-playground/validator 完全相同）：

```go
type User struct {
    FirstName      string     `validate:"required"`
    LastName       string     `validate:"required"`
    Age            uint8      `validate:"gte=0,lte=130"`
    Email          string     `validate:"required,email"`
    Gender         string     `validate:"oneof=male female prefer_not_to"`
    FavouriteColor string     `validate:"iscolor"`
    Website        string     `validate:"url"`
    Addresses      []*Address `validate:"required,dive,required"`
}

type Address struct {
    Street string `validate:"required"`
    City   string `validate:"required"`
    Planet string `validate:"required"`
    Phone  string `validate:"required"`
}
```

2. 生成校验代码：

```bash
quickvalidate -i ./path/to/your/package -o ./path/to/output/validate_gen.go
```

也可以在包里放一行 `//go:generate`：

```go
//go:generate quickvalidate -i . -o validate_gen.go
```

3. 调用生成的方法：

```go
user := &User{...}
if err := user.Validate(); err != nil {
    var errs errors.ValidationErrors
    if errors.As(err, &errs) {
        for _, e := range errs {
            fmt.Println(e.Namespace, e.Tag, e.Param, e.Kind)
        }
    }
}
```

```
User.Email  email    ""    string
User.Color  iscolor  ""    string   // e.ActualTag 是 "hexcolor|rgb|rgba|hsl|hsla"
User.Age    lte      130   uint8
```

## 命令行参数

- `-i`：输入文件或目录（目录会递归处理 `.go` 文件，自动跳过 `_test.go` 与 `*_gen.go`）
- `-o`：输出文件路径
- `-pkg`：包名（默认从输入推导）

## 支持范围

**已覆盖**：go-playground/validator v10.25.0 的全部内置 tag，包括

- 基础：`required`、`isdefault`、`omitempty`、`omitnil`、`omitzero`、`structonly`、`nostructlevel`
- 比较与长度：`len`、`min`、`max`、`eq`、`ne`、`gt`、`gte`、`lt`、`lte`、`eq_ignore_case`、`ne_ignore_case`（字符串按**字符数**，`time.Time` 与 `time.Duration` 有专门语义）
- 容器：`dive`、`keys`、`endkeys`、`unique`（切片/数组/map/按字段去重）
- 跨字段：`eqfield`、`nefield`、`gtfield`/`gtefield`/`ltfield`/`ltefield` 及其 `cs` 变体、`fieldcontains`、`fieldexcludes`
- 条件：`required_if/unless/with/with_all/without/without_all`、`excluded_*`、`skip_unless`
- 内容：`contains*`、`excludes*`、`startswith`、`endswith`、`startsnotwith`、`endsnotwith`
- 格式：`email`、`url`、`http_url`、`uri`、`urn_rfc2141`、`uuid*`、`ulid`、`ip/ipv4/ipv6/cidr`、`mac`、`hostname*`、`fqdn`、`json`、`jwt`、`base64*`、颜色（`hexcolor`、`rgb/rgba/hsl/hsla`、`iscolor`）、国家与币种（`iso3166_*`、`iso4217*`）、`bcp47_language_tag`、`cron`、`spicedb`、`credit_card`、`luhn_checksum`、`mongodb*`、`cve`、`semver`、`bic`、`dns_rfc1035_label`、`postcode_iso3166_*`、`eth_addr*`、`btc_addr*`、`isbn*`、`issn`、哈希（`md4/md5/sha*`、`ripemd*`、`tiger*`）、`*_addr` 解析类、`file/dir/filepath/dirpath/image`
- 别名：`iscolor`、`country_code`、`eu_country_code`
- 或分支：`rgb|rgba`、`oneof=a b|oneof=c d`

**不支持**：

- 跨包结构体的自动递归：只有同包内、同样由本工具生成的结构体才会被递归校验（validator 用反射可以穿透任意包）。这是唯一**静默跳过**的一类
- 一切**运行期注册**：自定义校验函数（`RegisterValidation`）、结构体级校验（`RegisterStructValidation`）、翻译器、`RegisterCustomTypeFunc`、`RegisterTagNameFunc`、`RegisterAlias`，以及 `WithRequiredStructEnabled` / `WithPrivateFieldValidation` / `SetTagName` 这些选项——生成出来的校验器在构建期就定死了，没有启动时可配的注册表
- 运行期入口 `Var` / `VarWithValue`（它们的 tag 是运行期字符串）与 `ValidateMap`
- `StructPartial` / `StructExcept` / `StructFiltered` 这类运行期选字段的入口

## 与 validator 的差异说明

生成代码以 `validator.New()`（默认配置）为对齐目标，因此：

- 非指针结构体字段上的 `required` 会被跳过——与 `requiredStructEnabled=false` 的默认行为一致
- `image` tag 对 `[]byte` 字段恒为 false——v10.25.0 的 `isImage` 只实现了字符串分支
- `iso3166_1_alpha_numeric*`、`iso4217_numeric`、`port` 按字段 kind 分派，与上游 `field.Kind()` 分支逐条对应：国家码接受字符串、有符号与无符号整数（先 `%1000` 再转型），币种码只接受整数（直接转型）。上游会 panic 的组合（如 `iso4217_numeric` + 字符串字段）在本工具中是**生成期错误**
- `spicedb` 的参数只接受 `permission`、`type`、`id` 或空，其它值在生成期报错（validator 在运行期 panic）
- `interface{}` 字段只能生成"是否已赋值"类的 tag（`required`、`isdefault`、`required_*`/`excluded_*` 家族、`omitempty`/`omitnil`/`omitzero`）；需要动态类型的 tag（`min`、`email`、`dive` 等）在**生成期报错**——validator 会按接口里的实际值在运行期校验，这是静态生成无法预先知道的
- 只能用于字符串的 tag 用在非字符串字段上（`email` 用在 `int`、`uuid` 用在结构体）是**生成期错误**：validator 会通过反射的 `String()` 读到占位串，于是这个字段会永远报失败
- `json` 是唯一同时接受字符串与 `[]byte` 的字符串类 tag（命名 `[]byte` 类型也可以）
- `timezone`、`bcp47_language_tag`、`*_addr` 等依赖运行时环境（时区库、DNS）的 tag，结果与运行环境相关，与 validator 一致
- 上游本身会 panic 的输入在差分测试中被标记跳过；当前只剩 `unique` 作用于含 nil 指针元素的切片一类（60 例，其余 5021 例均为真实比对）

## 对齐验证

`internal/align` 是差分测试套件：同一批结构体、同一批取值，分别跑真实库与本工具生成的代码，逐字段比对 `命名空间|tag|参数|ActualTag|StructNamespace|Kind`。

```bash
make align        # 或 go test ./internal/align/
```

覆盖方式：

- 手写边界用例（空值、边界值、Unicode、nil 指针、嵌入结构体、接口字段、经过指针的跨字段路径、map、别名、或分支、命名类型、整数回绕……）
- 随机差分测试：按字段类型从"有代表性取值池"中随机填充结构体并比对
- 179 个内置 tag **全部**出现在语料中；每次运行 5021 条比对，其中 60 条因上游 panic 跳过

## 性能

`go test -run '^$' -bench . ./examples/benchmark/`（本机实测，Apple M4 Pro）：

| 结构体 | go-playground/validator | 生成代码 |
|---|---|---|
| 扁平，6 个格式校验 | 631 ns，1 次分配 | 480 ns，0 次分配 |
| 嵌套，`required,dive,required` 遍历 100 个地址 | 10554 ns，101 次分配 | 2384 ns，0 次分配 |

倍数随结构体复杂度变化，建议用自己的结构体实测；`make benchmark` 仍会打印原来那个简单循环的结果。

## 构建与测试

```bash
make build      # 构建 CLI
make test       # 全部测试（含差分对齐套件）
make align      # 只跑差分对齐套件
make generate   # 重新生成示例与语料的校验代码
make examples   # 运行示例
make help       # 查看全部命令

go test -run '^$' -bench . ./examples/benchmark/   # 生成代码 vs 参考库
```

## 许可证

MIT。`pkg/validators` 中的校验实现与数据表移植自 [go-playground/validator](https://github.com/go-playground/validator)（MIT，Copyright (c) 2015 Dean Karn），文件头已注明出处。
