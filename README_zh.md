# QuickValidate

[English](README.md) | [中文](README_zh.md)

QuickValidate 是流行的 [go-playground/validator](https://github.com/go-playground/validator) 库的静态代码生成版本。它在构建时生成验证代码，而不是在运行时使用反射，这显著提高了性能。

## 特性

- 从结构体标签生成静态验证代码
- 兼容 go-playground/validator 的标签语法
- 比基于反射的验证快得多
- 没有运行时反射开销
- 类型安全的验证
- 编译时验证错误
- 支持常见的验证标签：required、email、url、oneof、min/max 等

## 安装

```bash
go install github.com/antlabs/quickvalidate/cmd/quickvalidate@latest
```

## 使用方法

1. 使用验证标签定义你的结构体（与 go-playground/validator 相同）

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

2. 运行 quickvalidate 命令生成验证代码

```bash
quickvalidate -i ./path/to/your/package -o ./path/to/output/validators.go
```

3. 使用生成的验证代码

```go
user := &User{...}
err := user.Validate() // 生成的方法
if err != nil {
    // 处理验证错误
}
```

## 性能

QuickValidate 比基于反射的验证快得多。基准测试显示，对于复杂的结构体，它可以快 10-20 倍。

## 命令行参数

- `-i`: 输入文件或目录路径（必需）
- `-o`: 输出文件路径（必需）
- `-pkg`: 包名（可选，默认从输入路径派生）

## 示例

项目中包含两个示例：

1. `examples/simple`: 一个简单的验证示例
2. `examples/benchmark`: 与 go-playground/validator 的性能比较

运行示例：

```bash
# 运行简单示例
make simple

# 运行基准测试
make benchmark
```

## 构建和测试

使用提供的 Makefile 可以轻松构建和测试项目：

```bash
# 构建二进制文件
make build

# 运行测试
make test

# 安装二进制文件
make install

# 查看所有可用命令
make help
```

## 许可证

MIT
