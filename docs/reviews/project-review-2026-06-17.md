 总体结论
  当前项目还没有达到“将 openapi.yml 转换为可用 WASM 模块”的闭环要求。OpenAPI 解析和 SDK 生成已经有雏形，但当前生成产物、运行时 facade、WASM-only 包和 TS 加载逻辑都存在阻断级问题。实测结果：

  go test ./... 通过。
  go run ./cmd/generator generate ... --oxlint-disable 可生成文件。
  生成后的 Go 包编译失败：undefined: runtime.IsValidEnum。
  标准 WASM 构建失败：undefined: runtime.ExportMain。
  GOOS=js GOARCH=wasm go test ./... 还暴露 pkg/runtime/convert 中 NewContextError、ErrCodeDeserializationFail 未定义。

  [必须修复] 生成后代码无法编译，WASM 构建闭环断裂
  建议：拆一个 pkg/runtime/runtime_wasm.go，加 //go:build js && wasm，只在 WASM 目标下 re-export wasm.ExportMain；普通 facade 里 re-export validate 函数；convert 包改成显式 runtimeerrors.NewContextError(...)。


  [建议修改] 当前 OpenAPI 模型是明显子集。pkg/generator/openapi.go:13 只覆盖 openapi/info/servers/paths/components.schemas，缺少全局 security、securitySchemes、path-level parameters、headers、cookies、callbacks、links、examples、encoding、style/explode 等字段。Schema 也没有 oneOf/anyOf/allOf/not/discriminator/
  default/min/max/pattern/minItems/maxItems 等约束。

  [必须修复] buildOperation 未判空 p.Schema，在参数没有 schema 或参数使用 $ref 时会 panic：pkg/generator/generator.go:336。
  [建议修改] $ref 仅支持 #/components/schemas/ 字符串截断：pkg/generator/generator.go:566，不解析参数、requestBody、response、外部文件、循环引用，也不校验引用目标是否存在。

  落地方案：优先引入 github.com/getkin/kin-openapi/openapi3 做加载、校验和 $ref 解析；内部生成模型只接收已 resolve 的 AST。复杂类型映射建议按阶段支持：先 allOf 合并结构体，再 oneOf/anyOf 生成 union wrapper 或 TS union，最后补 discriminator。

  2. WASM 性能与体积

  [建议修改] 当前已有 build/main.wasm 约 11 MB、examples/petstore/generated/main.wasm 约 12 MB，和 README 中标准 Go WASM 2-5 MB 的描述不一致。构建工具里的标准 Go 路径只用了 -trimpath：pkg/runtime/build/build.go:133，没有 -ldflags="-s -w"，也没有后处理压缩。

  建议：标准 Go 构建至少使用 -trimpath -ldflags="-s -w" -buildvcs=false；发布链路增加 wasm-opt -Oz 和 gzip/brotli 压缩指标。更重要的是默认优先 TinyGo，但要建立 TinyGo 兼容测试，因为当前依赖 net/http、syscall/js、反射和 vert，体积和兼容性都会受影响。若目标只是浏览器 HTTP SDK，建议把强类型和校验放 TS 侧，WASM 只保留必
  要业务逻辑，否则 Go runtime 成本不划算。

  3. 内存与数据交互

  [建议修改] JS/Go 交互大量依赖反射和逐字段对象转换：pkg/runtime/convert/converter.go:132、pkg/runtime/convert/converter.go:169。这会产生大量临时 map/slice，对大 body 或高频 API 调用不友好。jsValueToInterface 的普通对象分支没有过滤危险 key：pkg/runtime/convert/converter.go:153，而 JSValueToMap 有过滤，两者安全策略不
  一致。

  建议：跨边界统一传 Uint8Array 或 JSON string，Go 侧只做一次 json.Unmarshal；响应同理返回 []byte/string，由 TS 侧解析。对大响应，现在只用 io.LimitReader 读 10 MB：pkg/runtime/client/client.go:198，但无法判断是否被截断，建议读取 limit+1 字节并返回明确 response too large 错误。

  4. 工程质量与架构

  [必须修复] TS WASM 加载代码不符合 Go WASM 用法。pkg/generator/templates/sdk.ts.tmpl:98 把 window.Go 当实例使用，后面又写 new go.importObject() 和 new go.run(...)：pkg/generator/templates/sdk.ts.tmpl:103。正确形态应是：

  const go = new (window as any).Go();
  const result = await WebAssembly.instantiateStreaming(fetch(this.wasmUrl), go.importObject);
  go.run(result.instance);

  [建议修改] 生成的 APIClient.registerAll 只注册到实例 map：pkg/generator/templates/sdk.go.tmpl:134，但 WASM 导出层查的是全局 runtimeclient.GetOperation：pkg/runtime/wasm/exports.go:194。这导致 operationId 路由和生成 handler 没有接上。要么生成 init(){ runtime.RegisterOperation(...) }，要么让 WASMExports 持有生成
  client。

  5. 鲁棒性与边界处理

  [建议修改] ParseOpenAPI 直接 os.ReadFile 全量读入：pkg/generator/openapi.go:87，没有文件大小限制、schema 深度限制、引用循环检测。超大 spec 或恶意递归 $ref 后续一旦实现 resolve，会很容易造成内存/栈问题。

  [建议修改] 默认生成命令会跑 oxlint：cmd/generator/generate.go:105，我实测在当前网络受限环境下失败于 registry.npmjs.org。建议默认不联网安装工具：优先使用本地 node_modules/.bin/oxlint 或 PATH 中的 oxlint，找不到时给 warning，CI 再显式启用严格 lint。

  优先级建议

  先修编译闭环：facade re-export、convert 包错误引用、TS wasm loader、生成包 go test。
  再替换 OpenAPI 解析层为 kin-openapi，明确支持 OpenAPI 子集并加入 $ref resolve 测试。
  最后做体积优化和跨边界协议优化，目标是生成产物可编译、可运行、体积指标可复现。