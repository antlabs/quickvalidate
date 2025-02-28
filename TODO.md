# QuickValidate TODO List

以下是 QuickValidate 项目中尚未实现的功能点列表：

## 优先级高

### 数值比较验证器
- [x] `gt` (greater than) - 大于
- [x] `gte` (greater than or equal) - 大于等于
- [x] `lt` (less than) - 小于
- [x] `lte` (less than or equal) - 小于等于
- [x] `eq` (equal) - 等于

## 优先级中

### 字符串验证器
- [ ] `alpha` - 仅包含字母的字符串
- [ ] `alphanum` - 仅包含字母和数字的字符串
- [ ] `numeric` - 仅包含数字的字符串
- [ ] `hexcolor` - 十六进制颜色验证
- [ ] `rgb` - RGB颜色验证
- [ ] `rgba` - RGBA颜色验证

### 特殊格式验证器
- [ ] `uuid` - UUID格式验证
- [ ] `e164` - E.164电话号码格式验证
- [ ] `json` - JSON格式验证
- [ ] `datetime` - 日期时间格式验证

### 字符串内容验证器
- [ ] `contains` - 包含子字符串验证
- [ ] `startswith` - 以指定前缀开始验证
- [ ] `endswith` - 以指定后缀结束验证

## 优先级低

### 高级验证功能
- [ ] 结构体嵌套验证的扩展支持
- [ ] 自定义验证函数支持
- [ ] 条件验证支持 (required_if, required_unless 等)
- [ ] 跨字段验证 (eqfield, nefield 等)
- [ ] 自定义错误消息

## 注意事项
- 以上大部分验证器在 `pkg/validators/validators.go` 中已有实现，但在代码生成器 `pkg/generator/generator.go` 中尚未集成
- 数值比较验证器（gt, gte, lt, lte, eq）应优先实现，因为这些在示例中已经使用
