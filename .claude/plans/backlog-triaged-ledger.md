# go-swagger backlog — FIXED ledger

_Triaged issues that are **fixed / works-as-designed / N-A**, plus the single
verify-only row and the duplicates of fixed-or-active-fix issues. Split from the
first-pass triage (2026-06-16). Example snippets were stripped — re-pull the
issue body from GitHub if a concrete repro is needed; the per-issue **verdict**
(with its 📖 doc-site action where present) is retained below. Status legend and
columns are unchanged from the main tracker._

| # | Title | Summary | Example? | Status | Need doc |
|---:|-------|---------|:--------:|:------:|:------:|
| [91](https://github.com/go-swagger/go-swagger/issues/91) | provide extra spec generation comment annotations to fully decouple from source | Asks for comment annotations to declare query params pulled directly from http.Request, without defining a wrapper struct. | [▶](#a-91) | ✅ |  |
| [301](https://github.com/go-swagger/go-swagger/issues/301) | swagger:model huh? | User confused about what swagger:model does and when to use it; model not appearing in output. | [▶](#a-301) | ✅ | 📖 |
| [334](https://github.com/go-swagger/go-swagger/issues/334) | go-swagger not generating model info and showing error on swagger UI | generate spec produces empty definitions and Swagger UI errors for a simple REST service. | [▶](#a-334) | ✅ | 📖 |
| [361](https://github.com/go-swagger/go-swagger/issues/361) | Question about spec generation | How to annotate swagger:route/swagger:response for a particular response struct shape. | [▶](#a-361) | ♻️ | 📖 |
| [413](https://github.com/go-swagger/go-swagger/issues/413) | error reading embedded struct | 'unable to resolve embedded struct' error caused by an embedded custom (external) type. | [▶](#a-413) | ✅ |  |
| [439](https://github.com/go-swagger/go-swagger/issues/439) | Swagger too deep scans. Deep scan is causing the error. Expr (...) is unsupported for a schema | Scanner scans too deeply, parsing whole files and unrelated declarations, causing 'Expr unsupported for a schema'; suggests parsing only the specific discovered type. | [▶](#a-439) | ✅ |  |
| [452](https://github.com/go-swagger/go-swagger/issues/452) | swagger:model generation no import found for ... | 'no import found for null' / 'unknown primitive byte' when using guregu/null wrapper types in a model. | [▶](#a-452) | ✅ | 📖 |
| [559](https://github.com/go-swagger/go-swagger/issues/559) | Documentation / Tutorials? | Complains documentation on generating specs from annotations is too sparse; wants more code examples. |  | ✅ | 📖 |
| [610](https://github.com/go-swagger/go-swagger/issues/610) | Mimeheader primitive not found | 'mimeheader primitive not found' when using runtime.File in parameters. | [▶](#a-610) | ✅ | 📖 |
| [613](https://github.com/go-swagger/go-swagger/issues/613) | Unhelpful error message when unsupported types are used (was: Generating spec of external structs) | Unhelpful error when a referenced external struct is used; asks for implicit conversion of referenced types to models. | [▶](#a-613) | ✅ |  |
| [619](https://github.com/go-swagger/go-swagger/issues/619) | Response annotation missing type for interface | Response annotation omits the type for an interface field, producing an invalid response schema. | [▶](#a-619) | ✅ | 📖 |
| [622](https://github.com/go-swagger/go-swagger/issues/622) | Add support for response example objects | Feature request: support attaching example response objects to responses. | [▶](#a-622) | ✅ | 📖 |
| [624](https://github.com/go-swagger/go-swagger/issues/624) | Model responses don't get included in definitions | Model used as body in route responses isn't added to definitions (regression from #596); discovery/postDecls logic too tangled. | [▶](#a-624) | ✅ | 📖 |
| [779](https://github.com/go-swagger/go-swagger/issues/779) | Question: Adding reason for response message in swagger:route annotation | Asks how to set a response reason/description and required headers inline in swagger:route without extra structs. | [▶](#a-779) | ✅ | 📖 |
| [791](https://github.com/go-swagger/go-swagger/issues/791) | Add complete (design) example to documentation | Docs request: add a complete, commented design example explaining the spec-as-contract approach. | [▶](#a-791) | ✅ | 📖 |
| [796](https://github.com/go-swagger/go-swagger/issues/796) | Definitions are generated without use of `swagger:model` | Extra/unwanted definitions (param structs, interfaces) get generated with the -m flag even without swagger:model. | [▶](#a-796) | ✅ |  |
| [834](https://github.com/go-swagger/go-swagger/issues/834) | Generate swagger.json from existing code | Beginner generating spec from existing code: Format/Pattern/Type/Example fields seem ignored/concatenated into description; wants non-json tag naming. | [▶](#a-834) | ✅ | 📖 |
| [864](https://github.com/go-swagger/go-swagger/issues/864) | add begin-to-end tutorial about providing annotations to an existing code base | Docs request: a begin-to-end tutorial on annotating an existing codebase and keeping spec/client in sync. | [▶](#a-864) | ✅ | 📖 |
| [874](https://github.com/go-swagger/go-swagger/issues/874) | swagger:response doesn't work with package prefix | swagger:response fails to resolve a type referenced with a package prefix (utils.Error); likely a docs/parser bug. | [▶](#a-874) | ✅ | 📖 |
| [958](https://github.com/go-swagger/go-swagger/issues/958) | how to specify sample values in user structs | How to declare sample/example values in models; 'default' usage seems undocumented; default not carried for defined types. | [▶](#a-958) | ✅ |  |
| [977](https://github.com/go-swagger/go-swagger/issues/977) | How to use aliased string type as key in map? | Maps keyed by an aliased string type (map[Role][]int) aren't reflected properly in the generated spec. | [▶](#a-977) | ✅ |  |
| [989](https://github.com/go-swagger/go-swagger/issues/989) | add description to response | Wants description support on responses; setting text is misinterpreted as a model ref. | [▶](#a-989) | ✅ | 📖 |
| [1026](https://github.com/go-swagger/go-swagger/issues/1026) | Suggestion: x-logo feature | Feature request: support x-logo vendor extension for logo integration. | [▶](#a-1026) | ✅ | 📖 |
| [1063](https://github.com/go-swagger/go-swagger/issues/1063) | Making a Model readOnly | Wants a way to mark a model (field) as readOnly so it appears in GET but not POST bodies. | [▶](#a-1063) | ✅ | 📖 |
| [1078](https://github.com/go-swagger/go-swagger/issues/1078) | Fails to generate JSON schema for time fields | Time fields should be documented as strings in the generated JSON schema. |  | ✅ |  |
| [1079](https://github.com/go-swagger/go-swagger/issues/1079) | Don't generate schemas for models without swagger:model annotation | With -m, don't generate schemas for types lacking a swagger:model annotation (avoids junk schemas). | • | ✅ | 📖 |
| [1088](https://github.com/go-swagger/go-swagger/issues/1088) | editor error for array parameters. | Array parameters generate a structure rejected by the Swagger editor ('items $refs cannot match'). | [▶](#a-1088) | ✅ |  |
| [1092](https://github.com/go-swagger/go-swagger/issues/1092) | Error generating spec from main.go | 'mapping values are not allowed in this context' YAML error when generating spec from a package comment. | [▶](#a-1092) | ✅ |  |
| [1096](https://github.com/go-swagger/go-swagger/issues/1096) | Annotated go code fails to generate spec when using cgo | Annotated code that uses cgo yields an empty spec; works once cgo is removed. | [▶](#a-1096) | ✅ |  |
| [1109](https://github.com/go-swagger/go-swagger/issues/1109) | go-swagger generates invalid swagger.json when returning response is model | Invalid spec (missing definitions) when a route returns a swagger:model that has no corresponding swagger:response. | [▶](#a-1109) | ✅ | 📖 |
| [1115](https://github.com/go-swagger/go-swagger/issues/1115) | WARNING: Missing parser for a *ast.StarExpr | Type aliased to a pointer warns 'Missing parser for a *ast.StarExpr' and is omitted. | [▶](#a-1115) | ✅ |  |
| [1117](https://github.com/go-swagger/go-swagger/issues/1117) | How to define response as array type | How to declare a swagger:response whose body is an array type. | • | ✅ | 📖 |
| [1121](https://github.com/go-swagger/go-swagger/issues/1121) | Question: How to add description to tags? | How to add descriptions to tags (global tags section) when generating from code. |  | ✅ | 📖 |
| [1133](https://github.com/go-swagger/go-swagger/issues/1133) | Scan Models chokes on unsupported type | 'scan models chokes on unsupported type' for a function type alias; asks why unrelated types stop the scan. | [▶](#a-1133) | ✅ |  |
| [1174](https://github.com/go-swagger/go-swagger/issues/1174) | go-swagger fails with unsupported type when generating models although struct is not part of swagger spec | Scan fails on an unsupported function type even though it isn't part of the spec; asks to inspect only annotated fields. | [▶](#a-1174) | ♻️ |  |
| [1228](https://github.com/go-swagger/go-swagger/issues/1228) | Prop of type `map[string]interface{}` isn't added properly to model | A map[string]interface{} typed field is emitted as additionalProperties on the parent instead of as a property. | [▶](#a-1228) | ✅ |  |
| [1267](https://github.com/go-swagger/go-swagger/issues/1267) | Response files? | How to return a file (e.g. Excel) from an endpoint via annotations. | [▶](#a-1267) | ✅ | 📖 |
| [1279](https://github.com/go-swagger/go-swagger/issues/1279) | Add path parameters to swagger:route without using structs | Wants to declare path parameters inline in swagger:route without a wrapper struct. | [▶](#a-1279) | ✅ | 📖 |
| [1335](https://github.com/go-swagger/go-swagger/issues/1335) | generate spec: "can't determine selector path from runtime" | 'can't determine selector path from runtime' when a runtime.File field is used in parameters. | [▶](#a-1335) | ✅ | 📖 |
| [1350](https://github.com/go-swagger/go-swagger/issues/1350) | Generating empty spec | generate spec produces an empty spec with no errors; requests a verbose mode to trace scanned files. | [▶](#a-1350) | ✅ | 📖 |
| [1395](https://github.com/go-swagger/go-swagger/issues/1395) | error on converted yaml from generated json. | Definitions go wrong when converting the generated JSON to YAML via swaggerhub. | [▶](#a-1395) | ✅ | 📖 |
| [1398](https://github.com/go-swagger/go-swagger/issues/1398) | go-swagger spec generation fails due to ./vendor dir deps | Wants a way to exclude the vendor directory; vendored deps (lib/pq) fail the recursive scan. |  | ✅ | 📖 |
| [1401](https://github.com/go-swagger/go-swagger/issues/1401) | swagger file is never generated | Spec file never produced (only empty paths) when generating from tagged source; unsure if input is parsed. | [▶](#a-1401) | ✅ | 📖 |
| [1402](https://github.com/go-swagger/go-swagger/issues/1402) | additionalProperties breaks map[string]interface{} | additionalProperties on map[string]interface{} wrongly constrains values to object, breaking validation. | [▶](#a-1402) | ✅ |  |
| [1408](https://github.com/go-swagger/go-swagger/issues/1408) | Error on converted yaml from generated json | Duplicate of #1395: definitions corrupted when converting generated JSON to YAML; includes uploaded files. | [▶](#a-1408) | ♻️ |  |
| [1415](https://github.com/go-swagger/go-swagger/issues/1415) | Circular struct decencies casues erros in swagger spec | Circular struct dependency generates a spec that Swagger UI can't resolve ('Cannot read property of undefined'). | • | ♻️ |  |
| [1416](https://github.com/go-swagger/go-swagger/issues/1416) | Parameters do not get properly linked to operations. | Parameters defined via swagger:parameters don't get linked to the operation in the generated spec/UI. | [▶](#a-1416) | ✅ |  |
| [1436](https://github.com/go-swagger/go-swagger/issues/1436) | Weird generate spec behavior | swagger:model ignored without --scan-models, but with it unrelated structs are pulled in. | [▶](#a-1436) | ✅ | 📖 |
| [1459](https://github.com/go-swagger/go-swagger/issues/1459) | Example values on map keys instead of "AdditionalProp" | Wants to customize map example keys instead of the misleading 'additionalProp1-3'. | [▶](#a-1459) | ✅ | 📖 |
| [1483](https://github.com/go-swagger/go-swagger/issues/1483) | Items Doesn't support maps whilst using map[string]string | 'items doesn't support maps' when a field is map[string]string. | [▶](#a-1483) | ✅ | 📖 |
| [1499](https://github.com/go-swagger/go-swagger/issues/1499) | Not possible to override parameter schema type | Can't override a parameter field's schema type (e.g. expose Bar as string while keeping a richer Go type). | [▶](#a-1499) | ✅ |  |
| [1512](https://github.com/go-swagger/go-swagger/issues/1512) | Generating spec file from Go code can result with invalid integer formats (uint64) | uint64 emitted with format 'uint64', which is invalid in Swagger 2.0 and breaks generated clients. | • | ✅ | 📖 |
| [1542](https://github.com/go-swagger/go-swagger/issues/1542) | Inserting Examples in response schema | How to insert examples into a response schema. | [▶](#a-1542) | ✅ | 📖 |
| [1560](https://github.com/go-swagger/go-swagger/issues/1560) | Error on Mac OS X When Generating Spec | macOS cgo/GOPATH build failure (environmental); codescan scans cgo packages fine (witnessed #1096). | [▶](#a-1560) | ✅ |  |
| [1587](https://github.com/go-swagger/go-swagger/issues/1587) | Import detection fails if package path does not match package name | Import detection fails when package path doesn't match package name ('no import found for jose'); wants better error context. | [▶](#a-1587) | ✅ |  |
| [1595](https://github.com/go-swagger/go-swagger/issues/1595) | multi-line/block instead of single-line comments | Using /*...*/ block comments instead of // breaks YAML parsing of the spec annotations. | [▶](#a-1595) | ✅ |  |
| [1609](https://github.com/go-swagger/go-swagger/issues/1609) | Vendor extension x-example for Dredd not working | x-example vendor extension stripped to 'example'; needed for Dredd with required path params. | [▶](#a-1609) | ✅ | 📖 |
| [1613](https://github.com/go-swagger/go-swagger/issues/1613) | Generate spec for response with string content type | Generating spec for a response with content-type string yields no body / misplaced response. | [▶](#a-1613) | ✅ |  |
| [1635](https://github.com/go-swagger/go-swagger/issues/1635) | Spec response generation without inner struct | Response body ends up under 'headers' unless an inner struct is used; wants to avoid the inner struct. | [▶](#a-1635) | ✅ |  |
| [1665](https://github.com/go-swagger/go-swagger/issues/1665) | can't combine a same annotation across multiple-lines | Wants to spread one swagger:parameters annotation across multiple lines / reuse common params across many ops. | [▶](#a-1665) | ✅ |  |
| [1670](https://github.com/go-swagger/go-swagger/issues/1670) | Error operation (All): yaml: line 10: did not find expected key | YAML-body parse error tied to the Security-definition quirks; assumed resolved by the Security/grammar2 reworks (no repro to assert). | [▶](#a-1670) | ✅ |  |
| [1708](https://github.com/go-swagger/go-swagger/issues/1708) | Response annotation missing property type | 'missing property type' error for response structs; meta fields dropped (follow-up of #619). | [▶](#a-1708) | ✅ |  |
| [1711](https://github.com/go-swagger/go-swagger/issues/1711) | How to add a different description to parameters? | Wants a parameter description distinct from the x-go-name+description concatenation shown in UI. |  | ✅ | 📖 |
| [1713](https://github.com/go-swagger/go-swagger/issues/1713) | How to label the example Response & Request | Asks whether go-swagger supports example request/response labels in the spec. | [▶](#a-1713) | ✅ |  |
| [1725](https://github.com/go-swagger/go-swagger/issues/1725) | How to override  version with generate spec | How to inject the build-time version into the generated spec instead of hardcoding it in doc.go. | [▶](#a-1725) | ✅ | 📖 |
| [1726](https://github.com/go-swagger/go-swagger/issues/1726) | How to annotate lists? | Bullet-list markdown (*) in comments is lost in the generated description. | [▶](#a-1726) | ✅ | 📖 |
| [1727](https://github.com/go-swagger/go-swagger/issues/1727) | Is there a way to automatically generate the swagger spec from the code? | Asks whether the spec can be generated programmatically in-code rather than via the CLI. | [▶](#a-1727) | ✅ | 📖 |
| [1734](https://github.com/go-swagger/go-swagger/issues/1734) | How to avoid name conflicts between service and dependencies when generating spec | Name conflicts between a service and its dependencies (same struct name) produce nondeterministic specs. | [▶](#a-1734) | ♻️ | 📖 |
| [1735](https://github.com/go-swagger/go-swagger/issues/1735) | Golang 1.11 : "swagger generate spec" fail  > "assignment to entry in nil map" | 'assignment to entry in nil map' when generating spec after upgrading to Go 1.11. | [▶](#a-1735) | ✅ |  |
| [1737](https://github.com/go-swagger/go-swagger/issues/1737) | bson.ObjectId in swagger:response generates $ref | bson.ObjectId in a response generates a $ref that overrides the description. | [▶](#a-1737) | ✅ | 📖 |
| [1742](https://github.com/go-swagger/go-swagger/issues/1742) | Parameters | Can route parameters reference parameter structs located in a different package/location? | [▶](#a-1742) | ✅ | 📖 |
| [1758](https://github.com/go-swagger/go-swagger/issues/1758) | Newbie help: generate spec from source | Newbie: generate spec from source produces nothing, no error; project spread across packages. | [▶](#a-1758) | ✅ | 📖 |
| [1761](https://github.com/go-swagger/go-swagger/issues/1761) | Host url | Wants the host field to be dynamic per environment instead of manual edits. | [▶](#a-1761) | ✅ | 📖 |
| [1772](https://github.com/go-swagger/go-swagger/issues/1772) | Post Example | Asks for an example of passing body parameters in a POST. | [▶](#a-1772) | ✅ | 📖 |
| [1795](https://github.com/go-swagger/go-swagger/issues/1795) | Is it possible to generate a header with Basic base64(data) for a post request? | Asks how to document a Basic base64 Authorization header for a POST. | [▶](#a-1795) | ✅ | 📖 |
| [1815](https://github.com/go-swagger/go-swagger/issues/1815) | Is this possible to create/regenerate spec from swagger generated server files? | Can a spec be regenerated from server files previously generated by go-swagger? | [▶](#a-1815) | ✅ | 📖 |
| [1828](https://github.com/go-swagger/go-swagger/issues/1828) | generate/spec: route responses tag description | swagger:route response 'description' yields invalid spec with no description. | [▶](#a-1828) | ✅ |  |
| [1852](https://github.com/go-swagger/go-swagger/issues/1852) | Schema error for delete operation | Spec generated from the petstore fixture fails editor.swagger.io validation on a delete operation. | [▶](#a-1852) | ✅ | 📖 |
| [1865](https://github.com/go-swagger/go-swagger/issues/1865) | swagger:parameters split into new lines | A very long swagger:parameters operation-id line; asks how to split it across lines. | [▶](#a-1865) | ♻️ | 📖 |
| [1867](https://github.com/go-swagger/go-swagger/issues/1867) | PATCH operation | PATCH via swagger:operation silently fails; swagger:route works but can't specify the JSON body. | [▶](#a-1867) | ✅ |  |
| [1881](https://github.com/go-swagger/go-swagger/issues/1881) | Key words in comments | User can't find the full list of annotations; an array-of-object response yields invalid spec. | [▶](#a-1881) | ✅ | 📖 |
| [1887](https://github.com/go-swagger/go-swagger/issues/1887) | generate spec: file type support. | Cannot set Swagger type 'file' for parameters/models. | [▶](#a-1887) | ✅ | 📖 |
| [1891](https://github.com/go-swagger/go-swagger/issues/1891) | go-swagger doesn't work for group type ? | Group type declarations and single-file vs multi-file layout produce different/wrong specs. | [▶](#a-1891) | ✅ |  |
| [1913](https://github.com/go-swagger/go-swagger/issues/1913) | Sub-types not generated from go discriminated type | Sub-types of a discriminated type aren't emitted unless -m, which then over-generates response models. | [▶](#a-1913) | ✅ | 📖 |
| [1925](https://github.com/go-swagger/go-swagger/issues/1925) | parameters do not support interface | Interface-typed parameters abort spec generation. | [▶](#a-1925) | ✅ |  |
| [1931](https://github.com/go-swagger/go-swagger/issues/1931) | how to generate spec for package | How to generate a spec covering all sub-packages, not just doc.go. | [▶](#a-1931) | ✅ | 📖 |
| [1934](https://github.com/go-swagger/go-swagger/issues/1934) | model declared in function not picked up | A response struct declared inside a function isn't picked up (params are). | [▶](#a-1934) | ✅ |  |
| [1955](https://github.com/go-swagger/go-swagger/issues/1955) | swagger:parameters and swagger:operation in different go package, operation cannot use parameters | swagger:operation can't use swagger:parameters defined in a different package. | [▶](#a-1955) | ✅ | 📖 |
| [1958](https://github.com/go-swagger/go-swagger/issues/1958) | Ability to skip embedded tags | Wants to keep embedded vendor extensions (e.g. x-amazon-apigateway-integration) that nest under recognized tags like 'responses'. | [▶](#a-1958) | ✅ | 📖 |
| [1974](https://github.com/go-swagger/go-swagger/issues/1974) | Unable to run `swagger.go` | 'assignment to entry in nil map' running swagger.go generate spec; needs help. | [▶](#a-1974) | ✅ | 📖 |
| [2002](https://github.com/go-swagger/go-swagger/issues/2002) | Generate spec fails with invalid type error | Recent go-modules change broke generate spec: types in other packages no longer scanned ('invalid type'). | [▶](#a-2002) | ✅ |  |
| [2013](https://github.com/go-swagger/go-swagger/issues/2013) | swagger generate spec panic: runtime error: index out of range | panic: index out of range in generate spec when using go modules. | [▶](#a-2013) | ✅ |  |
| [2020](https://github.com/go-swagger/go-swagger/issues/2020) | Parameters not detected with multiple structs in one statement | Parameters not detected when multiple structs are declared in one type(...) block; annotation must be above 'type'. | [▶](#a-2020) | ✅ |  |
| [2027](https://github.com/go-swagger/go-swagger/issues/2027) | Embedded struct Support | Asks whether embedded struct (inheritance) support is planned (refers to long-open #413). | [▶](#a-2027) | ✅ | 📖 |
| [2038](https://github.com/go-swagger/go-swagger/issues/2038) | Swagger ignore json tags for embedded structures | Swagger ignores json tags for embedded structs, producing a nesting different from actual JSON output. | [▶](#a-2038) | ✅ | 📖 |
| [2062](https://github.com/go-swagger/go-swagger/issues/2062) | Cannot add security and SecurityDefinitions in swagger:operation | Can't add security/SecurityDefinitions in swagger:operation (works in route/meta). | [▶](#a-2062) | ✅ |  |
| [2064](https://github.com/go-swagger/go-swagger/issues/2064) | add example to string parameter in request body | example and default of a string body parameter are missing in the generated spec. | [▶](#a-2064) | ✅ |  |
| [2106](https://github.com/go-swagger/go-swagger/issues/2106) | `swagger generate spec` ignores `Extensions` on models when type is not an array | generate spec ignores Extensions on a model field unless the field is an array. | [▶](#a-2106) | ✅ |  |
| [2119](https://github.com/go-swagger/go-swagger/issues/2119) | Add flag to skip generation of `x-go-name` | Feature request: a flag to skip generating x-go-name / x-go-package (which clash across same-named types). | [▶](#a-2119) | ✅ | 📖 |
| [2125](https://github.com/go-swagger/go-swagger/issues/2125) | Parsing meta info comments can parse fields wrong. | Markdown content in swagger:meta is mis-parsed (e.g. Api-Version read as Version), ending comment parsing; proposes regex fix. | [▶](#a-2125) | ✅ |  |
| [2126](https://github.com/go-swagger/go-swagger/issues/2126) | Reference swagger models under a specific package | How to reference same-named models from different packages without mixing their examples. | [▶](#a-2126) | ♻️ | 📖 |
| [2127](https://github.com/go-swagger/go-swagger/issues/2127) | How to swagger ignore specific lines in the documentation? | Asks whether a single line (a vendor type reference) can be swagger:ignored. | [▶](#a-2127) | ✅ | 📖 |
| [2133](https://github.com/go-swagger/go-swagger/issues/2133) | spec generating schema and type | Path param 'type' is emitted twice (in schema and beside it), failing validation. | [▶](#a-2133) | ✅ |  |
| [2160](https://github.com/go-swagger/go-swagger/issues/2160) | Example of array of structs is 0 valued when generating spec | Example values for an array of structs come out as 0 in the generated spec/UI. | [▶](#a-2160) | ✅ | 📖 |
| [2172](https://github.com/go-swagger/go-swagger/issues/2172) | property comment does not get generated into swagger.yaml | A property comment is dropped when the field is a $ref; wants the comment kept on $ref fields. | [▶](#a-2172) | ✅ | 📖 |
| [2183](https://github.com/go-swagger/go-swagger/issues/2183) | Different spec produced for same codebase | Different specs produced from the same codebase due to a swagger-annotated type embedding another (map ordering). | [▶](#a-2183) | ✅ |  |
| [2184](https://github.com/go-swagger/go-swagger/issues/2184) | spec generation fails to replace $ref with indicated type when using swagger:type annotation on struct | swagger:type annotation on a struct is ignored; spec keeps a $ref instead of the indicated type (regression vs v0.17). | [▶](#a-2184) | ✅ |  |
| [2208](https://github.com/go-swagger/go-swagger/issues/2208) | Scanner tests are excluded from build ? | Scanner/scan tests are excluded from the build by a '// +build !go1.11' tag; suggests removing it. | [▶](#a-2208) | ✅ |  |
| [2210](https://github.com/go-swagger/go-swagger/issues/2210) | unknown field 'URL' in struct literal of type spec.ContactInfo | Build fails: unknown field URL/Name/Email in spec.ContactInfo/License (go-openapi/spec API mismatch). | [▶](#a-2210) | ✅ |  |
| [2218](https://github.com/go-swagger/go-swagger/issues/2218) | Unable to connect a go-swagger parameter to a route | swagger:parameters struct not linked to its route; query parameter missing from generated YAML. | [▶](#a-2218) | ✅ | 📖 |
| [2228](https://github.com/go-swagger/go-swagger/issues/2228) | how to give empty summary in swagger:route | Wants to suppress the auto-derived summary (first line) in swagger:route. | [▶](#a-2228) | ✅ | 📖 |
| [2230](https://github.com/go-swagger/go-swagger/issues/2230) | [Question] How to define an example in json.RawMessage field of a struct | How to define an example for a json.RawMessage field (currently shows as array of ints). | [▶](#a-2230) | ✅ | 📖 |
| [2232](https://github.com/go-swagger/go-swagger/issues/2232) | Unable to generate tags with spaces using the spec generation tool | Cannot create tags containing spaces in swagger:route. | [▶](#a-2232) | ✅ | 📖 |
| [2233](https://github.com/go-swagger/go-swagger/issues/2233) | Generate spec: Unable to find responses defined in other package using swagger:operation | swagger:operation can't find response models defined in another package ('$refs must reference a valid location'). | [▶](#a-2233) | ✅ |  |
| [2245](https://github.com/go-swagger/go-swagger/issues/2245) | how to write swagger:response schema that produces application/xml | How to annotate a swagger:response that produces application/xml instead of json. | [▶](#a-2245) | ✅ | 📖 |
| [2248](https://github.com/go-swagger/go-swagger/issues/2248) | Dealing with time.Duration in response's header | time.Duration in a response header errors ('type in body is required'); wants to force header type / define headers separately. | [▶](#a-2248) | ✅ |  |
| [2251](https://github.com/go-swagger/go-swagger/issues/2251) | Problems getting map with non-string keys serialized in spec | Maps with non-string keys inconsistently serialized (properties land in Headers or vanish). | [▶](#a-2251) | ✅ | 📖 |
| [2286](https://github.com/go-swagger/go-swagger/issues/2286) | Model accepted as Response | A swagger:model used directly as a swagger:response is accepted but per maintainer should require explicit body. | [▶](#a-2286) | ✅ | 📖 |
| [2294](https://github.com/go-swagger/go-swagger/issues/2294) | Unable to generate Swagger spec with more than one security header using "AND" logic. | Can't generate a spec requiring multiple security headers with AND logic. | [▶](#a-2294) | ✅ |  |
| [2296](https://github.com/go-swagger/go-swagger/issues/2296) | panic,embedded meet anonymous | panic when an embedded field meets an anonymous struct during schema build. | [▶](#a-2296) | ✅ |  |
| [2299](https://github.com/go-swagger/go-swagger/issues/2299) | Generated swagger schema is not deterministic | Generated schema is nondeterministic across runs (subtle field differences dirty the git tree). | [▶](#a-2299) | ✅ |  |
| [2305](https://github.com/go-swagger/go-swagger/issues/2305) | Query params and path params, enum dropdown | Query/path params produce malformed spec (duplicate description, illegal 'schema' on param, no enum dropdown). | [▶](#a-2305) | ✅ |  |
| [2311](https://github.com/go-swagger/go-swagger/issues/2311) | swagger:ignore documentation is incomplete with respect to fields. | Docs for swagger:ignore don't mention it also applies to struct fields (since #1497). | [▶](#a-2311) | ✅ | 📖 |
| [2317](https://github.com/go-swagger/go-swagger/issues/2317) | Extension x-nullable on a pointer has no effect | x-nullable extension on a pointer field has no effect (works only on arrays). | [▶](#a-2317) | ✅ |  |
| [2353](https://github.com/go-swagger/go-swagger/issues/2353) | Generation of valid spec file with body and request param | Generating a valid spec with both body and request params yields an invalid block that fails validate. | [▶](#a-2353) | ✅ |  |
| [2371](https://github.com/go-swagger/go-swagger/issues/2371) | Generate spec fails for response with array of objects | Spec generated from a server that was itself generated from a spec fails when a response is an array of objects. | [▶](#a-2371) | ✅ |  |
| [2379](https://github.com/go-swagger/go-swagger/issues/2379) | unsupported type "invalid type" error when using Linux binary | 'unsupported type "invalid type"' from the Linux v0.25 binary (not darwin) when referencing a schema in another package. | [▶](#a-2379) | ✅ |  |
| [2383](https://github.com/go-swagger/go-swagger/issues/2383) | Using inline struct inside of function that returns http.HandlerFunc can't generate spec | Inline structs inside a function returning http.HandlerFunc can't be found ('unable to find package and source file'). | [▶](#a-2383) | ✅ |  |
| [2384](https://github.com/go-swagger/go-swagger/issues/2384) | pattern containing '\n' is interpreted when generating comment producing illegal output. | A pattern containing '\n' is interpreted, producing illegal generated code; pattern should be a raw string. | [▶](#a-2384) | ✅ | 📖 |
| [2396](https://github.com/go-swagger/go-swagger/issues/2396) | model enum recognizion and handling spaces | enum: [a, b, c] vs enum: a, b, c produce different (wrong) results; expects a consistent enum array. | [▶](#a-2396) | ✅ |  |
| [2398](https://github.com/go-swagger/go-swagger/issues/2398) | Warn about duplicate definitions | Duplicate definitions (same id) silently override each other; wants a warning/error on overwrite. | [▶](#a-2398) | ♻️ | 📖 |
| [2403](https://github.com/go-swagger/go-swagger/issues/2403) | go swagger security auth0 | Global security (auth0) annotation isn't generated as expected from the doc comment. | [▶](#a-2403) | ✅ |  |
| [2407](https://github.com/go-swagger/go-swagger/issues/2407) | how add example to yml from golang code | example values from Go annotations don't appear in the generated YAML. | [▶](#a-2407) | ✅ |  |
| [2409](https://github.com/go-swagger/go-swagger/issues/2409) | Annotate structures with extensions | Can't annotate structs with Extensions to make a type import from its original package (related to #2106). | [▶](#a-2409) | ✅ | 📖 |
| [2412](https://github.com/go-swagger/go-swagger/issues/2412) | Cannot generate a parameter with type "file" | Cannot generate a parameter of Swagger type 'file' (Go has no file type; string breaks upload). | [▶](#a-2412) | ✅ | 📖 |
| [2417](https://github.com/go-swagger/go-swagger/issues/2417) | Embedding of aliased type | Embedding an aliased type from another package doesn't work. | [▶](#a-2417) | ✅ |  |
| [2419](https://github.com/go-swagger/go-swagger/issues/2419) | feature: support custom swagger type for struct field | Feature request: set a custom swagger:type on a struct field of an external library type (e.g. protobuf wrappers). | [▶](#a-2419) | ✅ |  |
| [2441](https://github.com/go-swagger/go-swagger/issues/2441) | File upload how to describe in annotations? | How to annotate a raw (non-multipart) file upload body (video/mp4, type string format binary). | [▶](#a-2441) | ✅ | 📖 |
| [2479](https://github.com/go-swagger/go-swagger/issues/2479) | How to disable security on a route but keep on all endpoints? | How to disable global security on one route while keeping it on the rest. | [▶](#a-2479) | ✅ |  |
| [2483](https://github.com/go-swagger/go-swagger/issues/2483) | Generation from code not working with allOf | Generating from code with allOf produces an invalid spec (duplicated values, wrong $ref). | [▶](#a-2483) | ✅ |  |
| [2520](https://github.com/go-swagger/go-swagger/issues/2520) | Generate spec fails with unsupported type "invalid type" | generate spec fails with 'unsupported type "invalid type"'. | [▶](#a-2520) | ✅ | 📖 |
| [2528](https://github.com/go-swagger/go-swagger/issues/2528) | go-swagger documentation is not updated for swagger enum | Docs don't explain swagger:enum usage. | [▶](#a-2528) | ✅ | 📖 |
| [2539](https://github.com/go-swagger/go-swagger/issues/2539) | Generate spec with additionalProperties | No way to set additionalProperties (e.g. additionalProperties:false) when generating spec from code — now delivered via explicit markers/keywords. | [▶](#a-2539) | ✅ | 📖 |
| [2547](https://github.com/go-swagger/go-swagger/issues/2547) | Unable to set a field value as empty string "" | Setting an example value to empty string '' fails (emits '' or escaped quotes). | [▶](#a-2547) | ✅ | 📖 |
| [2549](https://github.com/go-swagger/go-swagger/issues/2549) | Example not working for imported Types | example not applied for a field of an imported library type (shows 0). | [▶](#a-2549) | ✅ | 📖 |
| [2575](https://github.com/go-swagger/go-swagger/issues/2575) | Custom Request Headers | How to document custom request headers per the Swagger standard. | [▶](#a-2575) | ✅ | 📖 |
| [2588](https://github.com/go-swagger/go-swagger/issues/2588) | Panic on parsing interface type definition | panic when parsing a type definition that embeds an interface inside a struct. | [▶](#a-2588) | ✅ |  |
| [2592](https://github.com/go-swagger/go-swagger/issues/2592) | Wrap generated spec with additional payload | Wants to wrap generated responses with an extra middleware-added payload (e.g. a token object). | [▶](#a-2592) | ✅ | 📖 |
| [2596](https://github.com/go-swagger/go-swagger/issues/2596) | Overwrite Interface with specific type in response annotation | Generic envelope with an interface{} payload — document a concrete type per route via embed-and-shadow doc-only structs. | [▶](#a-2596) | ✅ | 📖 |
| [2599](https://github.com/go-swagger/go-swagger/issues/2599) | Custom type on models fields | Wants to override a model field's type when strfmt isn't usable (same as #2404). | [▶](#a-2599) | ✅ | 📖 |
| [2611](https://github.com/go-swagger/go-swagger/issues/2611) | How can I  generate a swagger spec without  Vendor Extensions? | Asks how to generate a spec without any vendor extensions (x-*). |  | ✅ | 📖 |
| [2618](https://github.com/go-swagger/go-swagger/issues/2618) | How to add a description to the fields in the body part of JSON type in the swagger API？ | Wants field descriptions in body JSON taken from struct tags/reflection. | [▶](#a-2618) | ✅ | 📖 |
| [2625](https://github.com/go-swagger/go-swagger/issues/2625) | How to generate spec from code if have only one struct for all responses? | How to generate spec when one response struct is reused for all endpoints with different embedded data. | [▶](#a-2625) | ✅ | 📖 |
| [2626](https://github.com/go-swagger/go-swagger/issues/2626) | Single line comment should never be parsed as title | A single-line struct comment ending in a period is parsed as title instead of description; now an opt-in option treats single-line comments as description. | [▶](#a-2626) | ✅ | 📖 |
| [2633](https://github.com/go-swagger/go-swagger/issues/2633) | The `swagger generate spec` command does not run normally. | Silent stop / empty spec from a degraded environment — now a loud diagnostic via the fail-loud feature. | [▶](#a-2633) | ✅ | 📖 |
| [2637](https://github.com/go-swagger/go-swagger/issues/2637) | Cyclic type definition for defined types using the same name in spec generation | Same-named types across packages create a cyclic $ref in the spec, then hang code generation. | [▶](#a-2637) | ✅ |  |
| [2638](https://github.com/go-swagger/go-swagger/issues/2638) | Improper handling of multiple variables on one line | Multiple variables declared on one line in a struct (e.g. color.RGBA's R,G,B,A) — only the first field is emitted. | [▶](#a-2638) | ✅ |  |
| [2639](https://github.com/go-swagger/go-swagger/issues/2639) | Generate schemas only for referenced Models | Wants --scan-models to emit only models actually referenced by $ref, not all in the library. | [▶](#a-2639) | ✅ | 📖 |
| [2651](https://github.com/go-swagger/go-swagger/issues/2651) | Wrong binding when swagger:operation uses parameters and swagger:parameters bind to operations at same time | Wrong parameter binding when an operation mixes inline parameters with swagger:parameters-bound ones. | [▶](#a-2651) | ✅ |  |
| [2652](https://github.com/go-swagger/go-swagger/issues/2652) | how to add complex example for a swagger:model | How to add a complex example for a swagger:model itself, and for fields whose complex type becomes a $ref (losing the example). | [▶](#a-2652) | ✅ |  |
| [2655](https://github.com/go-swagger/go-swagger/issues/2655) | Tags field ignored in Metadata when generating spec | Tags field in swagger:meta is ignored (so per-tag descriptions can't be set); Tags missing from meta fields list. | [▶](#a-2655) | ✅ |  |
| [2662](https://github.com/go-swagger/go-swagger/issues/2662) | Same ref names while generating spec | Same-named structs in different packages silently override each other producing an invalid spec; wants unique refs or an error. | [▶](#a-2662) | ♻️ |  |
| [2663](https://github.com/go-swagger/go-swagger/issues/2663) | How to document body parameter in a POST request ? | Doc question — the correct idiom is one `in: body` field whose Go type is the body schema (witnessed). | [▶](#a-2663) | ✅ | 📖 |
| [2687](https://github.com/go-swagger/go-swagger/issues/2687) | Ignore kubebuilder annotations | Wants to exclude kubebuilder-style '+' annotations from generated model/property descriptions. | [▶](#a-2687) | ✅ |  |
| [2701](https://github.com/go-swagger/go-swagger/issues/2701) | In path parameter for an embedded struct is ignored and thus default to  in query parameter | An 'in: path' parameter on an embedded struct is ignored and defaults to query. | [▶](#a-2701) | ✅ |  |
| [2746](https://github.com/go-swagger/go-swagger/issues/2746) | Strfmt: array of UUID | Array of UUID generates [0] instead of UUID examples; swagger:strfmt uuid only works for a single value. | [▶](#a-2746) | ✅ |  |
| [2761](https://github.com/go-swagger/go-swagger/issues/2761) | generate: allOf in response doesn't use $ref | swagger:allOf on an embedded struct lists fields as properties instead of generating a $ref. | [▶](#a-2761) | ✅ | 📖 |
| [2762](https://github.com/go-swagger/go-swagger/issues/2762) | codescan tests fail on go 1.18 | codescan unit tests fail on Go 1.18 because schema field ordering changed (tests assume a fixed order). | [▶](#a-2762) | ✅ |  |
| [2791](https://github.com/go-swagger/go-swagger/issues/2791) | swagger:parameters not working as defined in docs | Copy-pasting the docs' swagger:parameters example reproduces generation errors; works only without annotations. | [▶](#a-2791) | ✅ | 📖 |
| [2799](https://github.com/go-swagger/go-swagger/issues/2799) | swagger generator should keep the format of the API description. | Generator strips indentation in API/field descriptions; wants the original formatting preserved. | [▶](#a-2799) | ✅ |  |
| [2801](https://github.com/go-swagger/go-swagger/issues/2801) | Spec is not generated if generic struct declaration is not in the same file as swagger:parameters | Spec empty when a generic struct's declaration is in a different file than its swagger:parameters annotation. | [▶](#a-2801) | ✅ |  |
| [2802](https://github.com/go-swagger/go-swagger/issues/2802) | Error generating spec if swagger:model is a generic struct | Error generating spec when a swagger:model is a generic struct. | [▶](#a-2802) | ✅ |  |
| [2804](https://github.com/go-swagger/go-swagger/issues/2804) | Should swagger:parameters support map[string][]string? | Asks whether swagger:parameters should support map[string][]string (currently errors). | [▶](#a-2804) | ✅ |  |
| [2778](https://github.com/go-swagger/go-swagger/issues/2778) | Generating Empty Spec File | Empty spec from a permission/degraded load — now a loud diagnostic via the fail-loud feature. | [▶](#a-2778) | ✅ | 📖 |
| [2783](https://github.com/go-swagger/go-swagger/issues/2783) | Models get mixed when using structs from several packages | Same-named structs from several packages get their model definitions mixed up. | [▶](#a-2783) | ✅ |  |
| [2837](https://github.com/go-swagger/go-swagger/issues/2837) | Responses defined in routes break with go 1.19 formatting in 1.30.* | Responses defined in routes break under Go 1.19 comment formatting (gofmt turns '+' into '-'). | [▶](#a-2837) | ✅ |  |
| [2838](https://github.com/go-swagger/go-swagger/issues/2838) | can not generate swagger spec | Empty 3-line spec was a Go version / toolchain mismatch; resolved by the `toolchain` directive in go.mod. | [▶](#a-2838) | ✅ | 📖 |
| [2846](https://github.com/go-swagger/go-swagger/issues/2846) | Wrong format yaml format for enums outside body | Enums outside the body (in: formData/query) produce malformed YAML ('- description: \|4-'). | [▶](#a-2846) | ♻️ |  |
| [2860](https://github.com/go-swagger/go-swagger/issues/2860) | Models do not show fields from the struct. | Models generated with -m on the petstore fixture omit the struct's fields/required. | [▶](#a-2860) | ✅ |  |
| [2871](https://github.com/go-swagger/go-swagger/issues/2871) | Question: dynamic examples | Wants per-endpoint dynamic examples (e.g. endpoint URL) reusing one error model without many structs. | [▶](#a-2871) | ✅ | 📖 |
| [2872](https://github.com/go-swagger/go-swagger/issues/2872) | "ExternalDocs" are not generating the 2.0 spec on swagger:meta | ExternalDocs in swagger:meta isn't emitted in the 2.0 spec. | [▶](#a-2872) | ✅ |  |
| [2874](https://github.com/go-swagger/go-swagger/issues/2874) | go-swagger silently stops parsing `swagger:allOf` when `GOROOT` is not set | Silent stop on a GOROOT-unset degraded load (dup of #2633) — now a loud diagnostic via the fail-loud feature. | [▶](#a-2874) | ✅♻️ | 📖 |
| [2875](https://github.com/go-swagger/go-swagger/issues/2875) | AllOf member does not generate an external $ref object | An allOf member doesn't generate an external $ref object (unlike the docs example). | [▶](#a-2875) | ✅ |  |
| [2886](https://github.com/go-swagger/go-swagger/issues/2886) | Invalid memory address or nil pointer (SIGSEGV)  => swagger generate spec | SIGSEGV/nil pointer in generate spec, with no indication of the offending code location. | [▶](#a-2886) | ✅ |  |
| [2897](https://github.com/go-swagger/go-swagger/issues/2897) | with go1.20 swagger missing definitions and refs | Go 1.20 (debian-repo binary) yields missing definitions/refs; self-built binary works. | [▶](#a-2897) | ✅ |  |
| [2898](https://github.com/go-swagger/go-swagger/issues/2898) | Output of []any is changed in 0.30 so that it is interpreted as slice of strings | []any output changed in 0.30 to be interpreted as a slice of strings instead of objects. | [▶](#a-2898) | ✅ | 📖 |
| [2899](https://github.com/go-swagger/go-swagger/issues/2899) | Example not being added to schema for body string params | example not added to the spec for string parameters sent in the body. | [▶](#a-2899) | ✅ |  |
| [2907](https://github.com/go-swagger/go-swagger/issues/2907) | Go-Swagger not generating properties in yaml file | Generated YAML lacks the properties of a response object. | [▶](#a-2907) | ✅ |  |
| [2909](https://github.com/go-swagger/go-swagger/issues/2909) | Regular cannot generate swagger automatically | Route path with a regex segment ({id:[0-9]+}) can't be parsed by swagger:route. | [▶](#a-2909) | ✅ |  |
| [2912](https://github.com/go-swagger/go-swagger/issues/2912) | Problem of swagger generate ,about struct tag  "json/form" | Parameter names come from Go field names; wants form/json struct tag (e.g. form:"sort_key") honored for naming. | [▶](#a-2912) | ✅ | 📖 |
| [2917](https://github.com/go-swagger/go-swagger/issues/2917) | classifier: unknown swagger annotation "extendee" when importing github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options | 'unknown swagger annotation "extendee"' when importing grpc-gateway openapiv2 options; the annotation regex over-matches. | [▶](#a-2917) | ✅ |  |
| [2922](https://github.com/go-swagger/go-swagger/issues/2922) | [spec generation] enum description: superfluous name&values | String enum values are redundantly repeated in the generated description. | [▶](#a-2922) | ✅ | 📖 |
| [2924](https://github.com/go-swagger/go-swagger/issues/2924) | [question] is it possible to specify "x-go-type" extension when swagger spec generates? | Emit an x-go-type vendor extension on generated definitions — now available via an opt-in option. | [▶](#a-2924) | ✅ | 📖 |
| [2942](https://github.com/go-swagger/go-swagger/issues/2942) | how can i generate spec with a simple  string response | How to generate a spec for an endpoint returning a simple string response. | [▶](#a-2942) | ✅ |  |
| [2959](https://github.com/go-swagger/go-swagger/issues/2959) | Unable to provide security definitions | Including SecurityDefinitions in swagger:meta causes 'found character that cannot start any token' YAML error. | [▶](#a-2959) | ✅ |  |
| [2961](https://github.com/go-swagger/go-swagger/issues/2961) | Improper parsing of uint enums | uint enum values are incorrectly emitted as strings (uint types not handled in parseValueFromSchema). | [▶](#a-2961) | ✅ |  |
| [2963](https://github.com/go-swagger/go-swagger/issues/2963) | Unable to reference models in referenced module | Cross-module models via a `replace` directive resolve correctly (observed via local multi-module repro; no CI witness). | [▶](#a-2963) | ✅ | 📖 |
| [2980](https://github.com/go-swagger/go-swagger/issues/2980) | Should I expect embedded structs to be supported in generated spec files? | Embedded structs aren't typed in the generated spec; asks whether they're supported. | [▶](#a-2980) | ✅ | 📖 |
| [2985](https://github.com/go-swagger/go-swagger/issues/2985) | Need minAttributes and maxAttributes in the swagger:model annotation | code-to-spec lacks minProperties/maxProperties (minAttributes/maxAttributes) support that spec-to-code has. | [▶](#a-2985) | ✅ |  |
| [3005](https://github.com/go-swagger/go-swagger/issues/3005) | additionalProperties are lost when generating spec from code | additionalProperties are lost when scanning a generated model back with generate spec; proposes an annotation to preserve them. | [▶](#a-3005) | ✅ |  |
| [3007](https://github.com/go-swagger/go-swagger/issues/3007) | [Bug]generate spec error | 'strconv.ParseBool parsing "=false"' error caused by '+kubebuilder:default:=false' comments. | [▶](#a-3007) | ✅ | 📖 |
| [3013](https://github.com/go-swagger/go-swagger/issues/3013) | How to set a example value for array/string response type? | How to set an example value for an array/string response type. | [▶](#a-3013) | ✅ |  |
| [3035](https://github.com/go-swagger/go-swagger/issues/3035) | Example spec for swagger:response does not produce example output | The docs' swagger:response example doesn't produce the example output (schema/properties missing). | [▶](#a-3035) | ✅ |  |
| [3069](https://github.com/go-swagger/go-swagger/issues/3069) | Is there a way to change the representation of one parameter of the request object? | Wants to customize how a single request parameter (a custom-marshalled complex type) is represented in the spec. | [▶](#a-3069) | ✅ | 📖 |
| [3100](https://github.com/go-swagger/go-swagger/issues/3100) | `in: formData` in `swagger:route` annotation translates to nothing (`in` field is omitted) in the yaml spec file | 'in: formData' in swagger:route is dropped from the YAML; code only accepts 'form'; asks for docs/fix. | [▶](#a-3100) | ✅ |  |
| [3107](https://github.com/go-swagger/go-swagger/issues/3107) | No struct definition in swagger generate | generate spec -m omits the struct fields, emitting only an empty definition. | [▶](#a-3107) | ✅ |  |
| [3117](https://github.com/go-swagger/go-swagger/issues/3117) | Swagger spec generating type property along with schema references | Body parameter with a schema $ref also gets an invalid sibling 'type: object'. | [▶](#a-3117) | ✅ |  |
| [3119](https://github.com/go-swagger/go-swagger/issues/3119) | Can not declare normal field in properties of schema object | Using swagger:operation, can't declare a normal field in a schema's properties without an extra response struct. | [▶](#a-3119) | ✅ | 📖 |
| [3134](https://github.com/go-swagger/go-swagger/issues/3134) | How to Generate Swagger Specification for Versioned APIs in a Single Route File Using Go-Swagger? | Doc-only — per-version specs by scanning per package tree; covered by the doc-site (scoping-the-scan). | [▶](#a-3134) | ✅ | 📖 |
| [3138](https://github.com/go-swagger/go-swagger/issues/3138) | How To mark a field as deprecated? | Asks how to mark a request/response field as deprecated (like #2042). | [▶](#a-3138) | ✅ | 📖 |
| [3213](https://github.com/go-swagger/go-swagger/issues/3213) | [spec/parsing] Consider TypeSpec comments | Comments on TypeSpec nodes are ignored (only GenDecl considered); both should be parsed. | [▶](#a-3213) | ✅ |  |
| [3214](https://github.com/go-swagger/go-swagger/issues/3214) | [spec/parsing] Incomplete parsing of referenced typed primitives | Referenced typed primitives only get title/description parsed; an enum declaration is wrongly used as the description. | [▶](#a-3214) | ✅ |  |

## Verdicts

_Per-issue triage verdicts (example snippets removed; see GitHub for the original issue body)._

<a id="a-91"></a>

### #91 — provide extra spec generation comment annotations to fully decouple from source

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/91)

**✅ Works-as-designed (verified 2026-06-16).** A query parameter can be declared
INLINE in the swagger:operation YAML body (`parameters: - name: type, in: query,
type: string`) — no Go wrapper struct required, which is the source-decoupling
the reporter wanted. Locked by `fixtures/bugs/91` + `TestCoverage_Bug91` + golden.


<a id="a-301"></a>

### #301 — swagger:model huh?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/301)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** The model is emitted with full
fidelity: title/description split, `required:[id,name,login]`, `minimum:1`,
`minLength:3` (the spaced `min length:` form works), `format:email` (strfmt),
and a self-referential `$ref` array for `friends`. The reporter's "model not
appearing" was the `-m` (ScanModels) requirement — an unreferenced
`swagger:model` only shows up under `-m` (asserted both ways). Locked by
`fixtures/bugs/301` + `TestCoverage_Bug301` + golden. 📖 Doc-site: when does a
model appear (route-reachable vs `-m`).


<a id="a-334"></a>

### #334 — go-swagger not generating model info and showing error on swagger UI

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/334)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** The two failures the reporter
hit are gone: a `swagger:route` placed **inside a function body** (not the
function's doc comment) is now detected (the `/users` path is emitted), and the
model is produced — so `definitions` and `paths` are no longer empty (no more
"options.definition is required" Swagger UI error). Locked by `fixtures/bugs/334`
+ `TestCoverage_Bug334` + golden. 📖 Doc-site quirk: placing `// swagger:model`
**before** the doc text (annotation-first) drops the title/description — they are
captured only when the annotation comes last. Document "annotation goes last".


<a id="a-361"></a>

### #361 — Question about spec generation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/361)

**♻️ Duplicate of #2625 + 📖 (verified 2026-06-14).** Same question as #2625: a
generic `Response{Meta MetaResponse; Data interface{}}` envelope renders `meta`
as a `$ref` and `data` as an empty (any-type) schema. Distinct per-endpoint
payloads are documented via response composition
(`200: body:Envelope{Data: ConcreteType}`), not by one shared struct. No new
fixture (cross-linked to the #2625 generic-envelope recipe). 📖 see #2625.


<a id="a-413"></a>

### #413 — error reading embedded struct

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/413)

**✅ Fixed / lock (verified 2026-06-14).** The legacy "unable to resolve embedded
struct for: RankBy" error is gone. Embedding an external-package struct that
itself carries a custom-typed field (`maps.NearbySearchRequest` with a `RankBy`
field) resolves via `go/packages` type info: the embedded fields are promoted as
parameters and the custom `RankBy` type maps to its underlying `string`. Locked
by `fixtures/bugs/413/{maps,}` + `TestCoverage_Bug413` + golden.


<a id="a-439"></a>

### #439 — Swagger too deep scans. Deep scan is causing the error. Expr (...) is unsupported for a schema

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/439)

**✅ Fixed by grammar2 / lock (verified 2026-06-16).** An unrelated, unannotated
type with fields the scanner can't model (`chan`, `time.Time`) is left untouched
— only the discovered (annotated/referenced) type is emitted, with no "Expr
unsupported for a schema" error. The scanner no longer over-scans unrelated
declarations. Locked by `fixtures/bugs/439` + `TestCoverage_Bug439` + golden.


<a id="a-452"></a>

### #452 — swagger:model generation no import found for ...

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/452)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** "no import found for null" /
"unknown primitive byte" with guregu/null wrapper types was the legacy loader's
import resolution; modern `go/packages` resolves external imports. For a specific
shape, annotate the wrapper with `swagger:type` / `swagger:strfmt`. **Doc-site
action:** note external wrapper types resolve, plus the annotation override.


<a id="a-610"></a>

### #610 — Mimeheader primitive not found

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/610)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** "mimeheader primitive not
found" with `runtime.File` in parameters is the same legacy selector issue as
#1335; the supported file-upload mechanism is the `swagger:file` marker (#1887).
**Doc-site action:** document `swagger:file` for uploads.


<a id="a-613"></a>

### #613 — Unhelpful error message when unsupported types are used (was: Generating spec of external structs)

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/613)

**✅ Fixed / lock (verified 2026-06-16).** A referenced struct (here via a
`[]*Ulimit` field) is implicitly converted into its own definition and referenced
by `$ref` — no unhelpful "unsupported types" error. Locked by `fixtures/bugs/613`
+ `TestCoverage_Bug613` + golden.


<a id="a-619"></a>

### #619 — Response annotation missing type for interface

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/619)

**✅ Works-as-designed + 📖 (verified 2026-06-14).** Not a bug. An `interface{}`
field yields an empty schema (no `type`), which is valid Swagger 2.0 and means
"any" — it does not make the response invalid. Separately, the reporter's struct
has no `in:body` marker, so codescan treats its fields as **response headers**,
not a body schema (which is why they expected a body and saw something else);
for a body schema use an `in:body` Body field. No code. 📖 Doc-site: response
field semantics (body vs headers) and `interface{}` → any-schema; cross-link
#2625 / #361.


<a id="a-622"></a>

### #622 — Add support for response example objects

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/622)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** An `example:` on a response
body field is emitted on the response schema (`example: Bob`). The reporter's
`exampleValue:` was non-standard — `example:` is the keyword. Locked by
`fixtures/bugs/622` + `TestCoverage_Bug622` + golden. (Per-mime response
`examples:` are the separate §10 item.) **Doc-site action:** document `example:`
on response fields.


<a id="a-624"></a>

### #624 — Model responses don't get included in definitions

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/624)

The reporter's repro is the `booking` API ([bfirsh fork, api.go](https://github.com/bfirsh/go-swagger/blob/5ae46516fd61c0eb33365c1467eb4144c0bc5d03/fixtures/goparsing/bookings/api.go)) — a model (`DateRange`) used only as a field of a response body was not added to `definitions`.

**✅ Already fixed (verified 2026-06-12).** The model-discovery rule now works as the reporter expected and is locked by the existing fixture/test — no new fixture added (the example is already in-repo as `fixtures/goparsing/bookings/api.go`):

- `Customer` (`swagger:model`) → definition ✓
- `Booking` (external cross-repo struct, used in the body) → definition ✓
- `DateRange` (un-annotated, but referenced by the response body) → definition ✓
- `IgnoreMe` (un-annotated, unreferenced) → excluded ✓
- `BookingResponse` (`swagger:response`) → not a definition ✓

Asserted by `TestAppScanner_Definitions` (`internal/integration/petstore_test.go`), golden `bookings_spec.json`. The issue's two extra elements vs the in-repo fixture expose nothing new: a bare `// swagger:model` (name derived from the type) is covered across many fixtures, and `500: body:ErrorResponse` (a model body in a route response) is covered by #1109's `bodyref` fixture and the `routes-responses-*` fixtures.

📖 **Doc-site action:** document the model-discovery rule — a struct lands in `definitions` when it is `swagger:model`-annotated **or** referenced (transitively) by a route/response body; an un-annotated, unreferenced struct is excluded.


<a id="a-779"></a>

### #779 — Question: Adding reason for response message in swagger:route annotation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/779)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** An inline response
description in `swagger:route` is the `NNN: description:<text>` form (locked by
#1828). Richer per-response detail — headers, named schemas — uses a
`swagger:response` struct or the `swagger:operation` YAML body. **Doc-site
action:** show the inline `200: description:...` form alongside the
swagger:response form.


<a id="a-791"></a>

### #791 — Add complete (design) example to documentation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/791)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A complete, commented
design/spec-as-contract example is a pure documentation request — no scanner
behaviour. **Doc-site action:** a worked end-to-end example page (annotations →
spec → contract).


<a id="a-796"></a>

### #796 — Definitions are generated without use of `swagger:model`

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/796)

**✅ Already fixed (verified 2026-06-12).** Under `-m` (`ScanModels`) only
the `swagger:model` type (`pingResponse`) is emitted as a definition; the
un-annotated `Handler` interface and the `swagger:parameters` struct
(`pingParams`) are no longer leaked into `definitions`. The secondary bug
in snippet 3 (the `in: path` line folded into the `who` property
description) is also gone — the parameter renders with a clean description
and a real `in: path` field. Locked by
`internal/integration/coverage_bug_796_test.go` (fixture
`fixtures/bugs/796/`, golden `bugs_796_schema.json`).


<a id="a-834"></a>

### #834 — Generate swagger.json from existing code

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/834)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** `pattern` / `max length` /
`example` on a field are applied as schema fields — NOT concatenated into the
property description. Locked by `fixtures/bugs/834` + `TestCoverage_Bug834` +
golden. The reporter's other ask — naming a property from a non-`json` struct
tag — is forthcoming-features §7 (#1391). **Doc-site action:** document the field
validation/example keywords and link the non-json naming-tag plan.


<a id="a-864"></a>

### #864 — add begin-to-end tutorial about providing annotations to an existing code base

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/864)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A begin-to-end tutorial on
annotating an existing codebase (and keeping spec/client in sync) is a
documentation request. **Doc-site action:** a "retrofit annotations onto an
existing API" tutorial; pairs with #791's worked example and the
empty-output/scan-scope guidance (#1758 etc.).


<a id="a-874"></a>

### #874 — swagger:response doesn't work with package prefix

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/874)

**✅ Fixed (warn) / 📖 Need doc (2026-06-12).** The OP's underlying goal — a
response type living in another package — already works: cross-package
types resolve automatically and a **plain** response name (`utilsError`,
not `utils.Error`) round-trips fine. The package-prefix dot was a
misunderstanding: response/definition names are JSON labels, not
Go-qualified identifiers (we deliberately keep the name alphabet to plain
identifiers — opening it to `.` invites arbitrary characters).

The real defect was UX: a dotted name (`swagger:response utils.Error`) was
**silently dropped** — the loose classifier accepted the marker, the strict
name matcher rejected the name, and the annotation produced nothing with no
feedback. Fixed: codescan now emits a `WARNING` when a `swagger:model` /
`swagger:response` name is not a plain identifier, instead of dropping it
silently (`internal/scanner/index.go#warnMalformedStructName`,
`parsers.MalformedModelName` / `MalformedResponseName`, unit test
`TestMalformedOverrideName`). The route-side ref (`200: utils.Error`)
already emits its own "not found … dropped" diagnostic (see #1109).

📖 **Doc-site action:** document that response/definition names are plain
labels (not Go-qualified); cross-package types are resolved automatically,
so use `Error`, never `utils.Error`.


<a id="a-958"></a>

### #958 — how to specify sample values in user structs

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/958)

**✅ Already fixed (verified 2026-06-12).** Field-level `default:` and
`example:` are carried both for plain builtin fields (inline on the
property) and for named/defined-type fields (snippet 2 — the reported
gap). For a defined-type field the keywords ride the override arm of an
allOf compound (a $ref cannot carry sibling keywords; same draft-4-shape
as #3125). Locked by `internal/integration/coverage_bug_958_test.go`
(fixture `fixtures/bugs/958/`, golden `bugs_958_schema.json`), covering
both default and example on plain and defined-type fields.


<a id="a-977"></a>

### #977 — How to use aliased string type as key in map?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/977)

**✅ Fixed / lock (verified 2026-06-16).** A map keyed by a named string type
(`type Role string`) renders as `{type:object, additionalProperties:<value>}` —
its underlying `string` kind makes it a valid object key. Locked by
`fixtures/bugs/977` + `TestCoverage_Bug977` + golden. (Integer-kind keys are the
separate §18 gap, #2251.)


<a id="a-989"></a>

### #989 — add description to response

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/989)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** The inline description text is no
longer misread as a model $ref. Inline response descriptions are fully captured
via the swagger:operation YAML body (403 → "Unauthorized"); the legacy
swagger:route Responses block maps a code to a response name and drops an inline
description value there. Locked by `fixtures/bugs/989` + `TestCoverage_Bug989` +
golden. 📖 Doc-site: how to give a response a description (swagger:operation or a
swagger:response object) — route-block inline descriptions are not captured.


<a id="a-1026"></a>

### #1026 — Suggestion: x-logo feature

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1026)

**✅ Works-as-designed / lock + 📖 (verified 2026-06-14).** No dedicated x-logo
knob is needed: arbitrary x-* vendor extensions are declared in swagger:meta via
`InfoExtensions:` (info-scoped) or `Extensions:` (spec root). x-logo lands on the
info object. Locked by `fixtures/bugs/1026` + `TestCoverage_Bug1026` + golden.
📖 Doc-site: document InfoExtensions / Extensions for custom vendor extensions.


<a id="a-1063"></a>

### #1063 — Making a Model readOnly

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1063)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** A `// read only: true` field
annotation emits `readOnly: true` on the property. Excluding the field from
request bodies is a downstream code-generation concern, not the spec scanner's.
Locked by `fixtures/bugs/1063` + `TestCoverage_Bug1063` + golden. 📖 Doc-site:
the readOnly annotation and its spec-level (not enforcement) semantics.


<a id="a-1088"></a>

### #1088 — editor error for array parameters.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1088)

**✅ Fixed (`fix/go-swagger-1088`).** Root cause: the SimpleSchema "catch at
exit" validator (`schema.validateSimpleSchemaOutcome`) only inspected the
top-level target. For an array element the `$ref` lives one level down, on
`param.Items` / `header.Items`, whose adapter `resolvers.ItemsTypable` did not
implement `schema.SimpleSchemaProbe` — so the validator skipped it and the
illegal `items: {$ref}` survived.

Fix: make `ItemsTypable` a `SimpleSchemaProbe` (`SimpleSchemaShape` / `HasRef`
/ `ResetForViolation`). Each array element is built through a fresh
`WithSimpleSchema` sub-build whose target IS the `ItemsTypable`, so the existing
validator now inspects the element shape and dissolves an illegal `$ref` to an
empty `{}` with a `CodeUnsupportedInSimpleSchema` diagnostic. Named primitives
(`[]Label`) keep expanding inline to `{type: string}`. Because `ItemsTypable` is
shared by parameters AND response headers, both are fixed by the one change.

Also added `common.Builder.ResetPostDeclarations`, called by the validator on a
ref violation: `MakeRef` had discovered the now-dissolved target's decl, which
would otherwise linger as an orphan definition (`[]Ele` no longer drags an
unreferenced `Ele` into `definitions`). This also removed a pre-existing orphan
in the `in-case-insensitive` golden (an `io.Reader` formData param that resets).

Witnesses: `fixtures/bugs/1088` (query `[]Label`/`[]Ele` + response headers
`X-Tags []Label`/`X-Objs []Ele`); `TestCoverage_Bug1088` asserts no `$ref` on
any items level, the empty-object dissolution, the absent orphan, and the
diagnostics. Golden drift (10 files) is uniformly the same correct dissolution
of array-of-object simple-schema items plus the one orphan removal. Documented
in `schema/README.md` §simple-schema-mode and `common/README.md` §postdecls.


<a id="a-1092"></a>

### #1092 — Error generating spec from main.go

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1092)

**✅ Fixed by grammar2 / lock (verified 2026-06-16).** A `swagger:meta` block with
colon-bearing prose (`host:port`, ratios like `3:1`, URLs) parses cleanly: the
prose stays in the info description, the real `Version:` field is read, and no
YAML "mapping values are not allowed in this context" error is raised. The legacy
engine mis-read colon-bearing prose as YAML mappings. Locked by
`fixtures/bugs/1092` + `TestCoverage_Bug1092` + golden.


<a id="a-1096"></a>

### #1096 — Annotated go code fails to generate spec when using cgo

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1096)

**✅ Fixed / lock (verified 2026-06-14).** A package importing "C" (cgo) used to
yield an empty spec; it now scans normally via go/packages (route + response
emitted). Locked by `fixtures/bugs/1096` + `TestCoverage_Bug1096` + golden, both
guarded by the `cgo` build tag so they skip where cgo is unavailable (keeps
cross-platform CI green).


<a id="a-1109"></a>

### #1109 — go-swagger generates invalid swagger.json when returning response is model

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1109)

Topic:

For such app it generates swagger.json without any definitions block and references to non-existing #/responses/order.

After after declaring such struct:

and using it in different route, swagger generates definitions block and references response correctly.

Reply from @casualjim (original author of this project):
> this is the expected behavior.
> If you want to skip the response struct you should use:

With this the struct will be detected as not used so so make it find the Order struct you have to run swagger generate spec -m ....
the -m will make it scan for swagger:model and include those structs

**✅ Already fixed / 📖 Need doc (verified 2026-06-12).** Not a codegen bug —
@casualjim's reply is correct. The original invalid-spec symptom (a dangling
`$ref: #/responses/order`) no longer occurs. `200: <name>` resolves in this
order: (1) a `swagger:response` named `<name>` → `$ref: #/responses/<name>`;
(2) otherwise a scanned model definition named `<name>` (requires `-m`) →
promoted to a body `$ref: #/definitions/<name>` (the OP's `200: order` case
now works directly with `-m`, more forgiving than the 2017 guidance); (3)
otherwise the ref is dropped with a warning diagnostic rather than emitting an
invalid reference. casualjim's `200: body:order` form also works. Both working
forms locked by `internal/integration/coverage_bug_1109_test.go` (fixtures
`fixtures/bugs/1109/{responseref,bodyref}`, goldens
`bugs_1109_responseref_schema.json` / `bugs_1109_bodyref_schema.json`).

📖 **Doc-site action:** explain the `-m` requirement for model discovery and
the `200: <name>` resolution order (named response vs model-definition
fallback vs `body:<name>`).


<a id="a-1115"></a>

### #1115 — WARNING: Missing parser for a *ast.StarExpr

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1115)

**✅ Fixed / lock (verified 2026-06-14).** A type aliased to a pointer
(`type BarResponse *Thing`) no longer warns "Missing parser for a
*ast.StarExpr" and is no longer omitted: grammar2 handles the StarExpr,
dereferences the pointer, and emits a `$ref` like the non-pointer alias. Locked
by `fixtures/bugs/1115` + `TestCoverage_Bug1115` + golden.


<a id="a-1133"></a>

### #1133 — Scan Models chokes on unsupported type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1133)

**✅ Fixed / lock (verified 2026-06-14).** An unsupported (function) type in a
scanned package no longer halts the scan: it is warned and skipped, and the
annotated model is still emitted. (`go-i18n`'s `TranslateFunc` is the original
trigger.) Locked by `fixtures/bugs/1133` + `TestCoverage_Bug1133` + golden;
covers #1174 too.


<a id="a-1174"></a>

### #1174 — go-swagger fails with unsupported type when generating models although struct is not part of swagger spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1174)

**✅ Fixed / lock ♻️ duplicate of #1133 (verified 2026-06-14).** Same scenario
as #1133: an unsupported function type used by a non-annotated struct, scanned
under `-m`, no longer fails — it is warned and skipped while the annotated model
is emitted. Folded into `fixtures/bugs/1133` + `TestCoverage_Bug1133` (the test
asserts the func type and its user struct are absent and the model present).


<a id="a-1228"></a>

### #1228 — Prop of type `map[string]interface{}` isn't added properly to model

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1228)

**✅ Works-as-designed (verified 2026-06-16).** A `map[string]interface{}` field
is emitted as a property with `{type:object, additionalProperties:{}}` — NOT
hoisted onto the parent model. Same behaviour locked by #1402
(`fixtures/bugs/1402`). The reporter's "not added properly" no longer reproduces.


<a id="a-1267"></a>

### #1267 — Response files?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1267)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A file-download response is
described with `produces: application/octet-stream` plus a body field marked
`swagger:strfmt binary` → `{type:string, format:binary}`. Locked by
`fixtures/bugs/1267` + `TestCoverage_Bug1267` + golden. (The bare `// format:
binary` line is prose on a body field; `swagger:strfmt binary` is the override.)
**Doc-site action:** document the binary-response recipe.


<a id="a-1279"></a>

### #1279 — Add path parameters to swagger:route without using structs

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1279)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** Path parameters can be declared
inline in a `swagger:route` comment via a `Parameters:` block (e.g. `+ name:
country / in: path / type: string / required: true`) — no wrapper
`swagger:parameters` struct required. Locked by `fixtures/bugs/1279` +
`TestCoverage_Bug1279` + golden. 📖 Doc-site: document the inline Parameters:
block in swagger:route.


<a id="a-1335"></a>

### #1335 — generate spec: "can't determine selector path from runtime"

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1335)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** The "can't determine selector
path from runtime" error for a `runtime.File` parameter was the legacy scan
engine's import-resolution; modern `go/packages` resolves cross-package imports,
and the supported file-upload mechanism is the `swagger:file` marker on a
formData field (locked by #1887 / #2412). **Doc-site action:** document
`swagger:file` for uploads (no special `runtime.File` handling needed).


<a id="a-1350"></a>

### #1350 — Generating empty spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1350)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** An empty spec with no errors
means no `swagger:*` annotations were found (or the wrong scan scope) — the spec
is annotation-driven, not derived from framework routes. `Options.Debug` gives
verbose tracing of scanned packages. Same guidance as #1758 / #1401. **Doc-site
action:** the getting-started "empty output ⇒ no annotations / scan ./..." note,
plus Debug.


<a id="a-1395"></a>

### #1395 — error on converted yaml from generated json

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1395)

**✅ Fixed / lock (re-triaged 2026-06-17).** Mis-titled ("error on converted
yaml from generated json") and earlier mis-shelved as an external-converter
wont-fix; per the issue thread (@casualjim) the real cause was a **misclassified
security entry** in the tab-indented `swagger:meta`. The OP's exact meta block now
parses cleanly into a correct `security` requirement + apiKey
`securityDefinitions` — no corrupted definitions. (The OP's `SecurityDefinition`
singular was a keyword typo.) Locked by `fixtures/bugs/1395` +
`TestCoverage_Bug1395` + golden; resolved by the meta-Security parsing work (cf.
#2403).



<a id="a-1401"></a>

### #1401 — swagger file is never generated

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1401)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** "Never generated / only empty
paths" is the annotation-driven / scan-scope situation: the spec is built from
`swagger:*` annotations across the scanned packages (`./...`), not from tagged
source alone. Same guidance as #1350 / #1758; `Options.Debug` traces what was
scanned. **Doc-site action:** covered by the empty-output getting-started note.


<a id="a-1402"></a>

### #1402 — additionalProperties breaks map[string]interface{}

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1402)

**✅ Fixed / lock (verified 2026-06-16).** `map[string]interface{}` renders as
`{type:object, additionalProperties:{}}` — an OPEN value schema (any), not one
wrongly constrained to object. Locked by `fixtures/bugs/1402` +
`TestCoverage_Bug1402` + golden.


<a id="a-1408"></a>

### #1408 — Error on converted yaml from generated json

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1408)

**♻️ Duplicate of #1395 (re-triaged 2026-06-17).** Same report as #1395, which
is now ✅ fixed (the misclassified meta security entry parses correctly). Moved
to the fixed ledger alongside #1395.



<a id="a-1416"></a>

### #1416 — Parameters do not get properly linked to operations.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1416)

**✅ Fixed / lock (verified 2026-06-14).** The reporter saw a spurious top-level
`$ref` emitted alongside the `schema` on a body parameter (malformed, breaking
the operation link). The body param is now clean — `{in: body, schema: {$ref}}`,
no top-level `$ref`. The operationId already matched, so linking works. Locked
by `fixtures/bugs/1416` + `TestCoverage_Bug1416` + golden.


<a id="a-1436"></a>

### #1436 — Weird generate spec behavior

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1436)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** Without `--scan-models`
only route-reachable models are emitted; with `-m` every `swagger:model` is
(standalone ones included). The reporter's observation is the intended `-m`
semantics. Locked by `fixtures/bugs/1436` + `TestCoverage_Bug1436`. The
"emit `-m` but prune unreferenced" middle ground is forthcoming-features §12
(#2639). **Doc-site action:** document `-m` vs route-reachable discovery.


<a id="a-1459"></a>

### #1459 — Example values on map keys instead of "AdditionalProp"

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1459)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A map field accepts an
`example` object with meaningful keys, carried on the additionalProperties
schema (e.g. `example: {timeout: 30s, retries: 3}`); locked by `fixtures/bugs/1459`
+ `TestCoverage_Bug1459` + golden. The misleading `additionalProp1/2/3` keys are
the Swagger UI's fallback rendering when no example is set — a UI artifact, but
the user's goal (meaningful example keys) IS controllable in the spec via
`example:`. **Doc-site action:** document setting an `example` on a map field to
override the UI's placeholder keys.


<a id="a-1483"></a>

### #1483 — Items Doesn't support maps whilst using map[string]string

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1483)

**✅ Fixed / lock + 📖 (verified 2026-06-16).** A `map[string]string` field in a
body parameter renders as `{type:object, additionalProperties:{type:string}}` —
no "items doesn't support maps" error. Locked by `fixtures/bugs/1483` +
`TestCoverage_Bug1483` + golden. (A map is only valid in an object/body context,
not as a bare query/path parameter.)


<a id="a-1499"></a>

### #1499 — Not possible to override parameter schema type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1499)

**✅ Fixed / lock (landed 2026-06-17, branch `fix/go-swagger-1499` → merged `3c7b7b6`).**
A `swagger:type` override on a `swagger:parameters` field whose Go type is a
struct is now honoured by the parameters builder — the parameter is emitted with
the overridden simple type instead of coming out **typeless** (invalid Swagger
2.0). The fix (`0883778`) brings the parameters builder in line with the
schema/model builder, which already honoured field-level `swagger:type`
(#2419/#2184). The RED witness `fixtures/bugs/1499` + `TestCoverage_Bug1499`
(`dbc10a3`) is now green. (The bare `type:` form without the `swagger:` prefix
stays prose, not an override.)


<a id="a-1542"></a>

### #1542 — Inserting Examples in response schema

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1542)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** An `example:` on a scalar field
is applied, and an `example:` on a map field is applied as an **object** —
provided the map example is written as valid JSON (`{"k":"v"}`). Locked by
`fixtures/bugs/1542` + `TestCoverage_Bug1542` + golden. 📖 Doc-site: complex
(map/object) examples must be valid JSON on the `example:` line (comma-list /
array coercion remains the §2.1 enhancement).


<a id="a-1560"></a>

### #1560 — Error on Mac OS X When Generating Spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1560)

**✅ Fixed / works-as-designed (re-triaged 2026-06-17).** The 2018 macOS
cgo/GOPATH toolchain failure building the custom-server example (crypto/x509 cgo
+ unresolved GOPATH imports) is environmental — not a codescan logic bug; modern
go/modules + go/packages loading do not hit it. codescan scans **cgo** packages
correctly, locked by `fixtures/bugs/1096` + `TestCoverage_Bug1096` (a
`//go:build cgo` package with `import "C"` + a `C.malloc` call). Cross-referenced
by commit `0dd122b` (`* contributes go-swagger#1560`). No further action.


<a id="a-1587"></a>

### #1587 — Import detection fails if package path does not match package name

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1587)

**✅ Fixed / lock (verified 2026-06-14).** A field referencing a type from a
package whose import-path tail (`josev2`) differs from its package name (`jose`)
used to fail with "no import found for jose". go/packages resolves it by the
real package name; the `$ref` and its definition are emitted. Locked by
`fixtures/bugs/1587` + `TestCoverage_Bug1587` + golden.


<a id="a-1595"></a>

### #1595 — multi-line/block instead of single-line comments

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1595)

**✅ Fixed (`fix/go-swagger-1595`).** Root cause: `parsers.parsePathAnnotation`
(shared by `swagger:operation` and `swagger:route`) reshapes the annotation's
comment into per-line synthetic `//` comments for `Remaining`. A `/* */` block
comment arrives as a single `*ast.Comment` whose `Text` still carries the
framing, so splitting it by `\n` turned the closing `*/` into a synthetic
`// */` line — which lands in the YAML body as `*/`, where the leading `*`
reads as a YAML alias indicator ("did not find expected alphabetic or numeric
character"), failing the whole scan.

Fix: in `parsePathAnnotation`, strip the `/* */` framing (`stripBlockFraming`)
before splitting, and shed the godoc `* ` continuation decoration per line
(`stripBlockContinuation`, mirroring grammar's, preserving YAML indentation
when no `*` is present). This handles both the flush-left block style (the
reporter's) and the idiomatic `*`-decorated style, and — since the helper is
shared — fixes `swagger:route` block comments too. No grammar-layer change; the
duplication of the tiny continuation-strip keeps the scanner-level `parsers`
package free of a dependency on the `grammar` sub-package.

Witnesses: `fixtures/bugs/1595` (flush-left + `*`-decorated `swagger:operation`)
and `TestCoverage_Bug1595`. No golden drift (block-comment annotations were
previously unrepresentable, so no fixture exercised the path). New helpers
documented via godoc.

gofmt-robustness verified (the F7 / 94ec08f quirk): the fixture is committed
gofmt-clean, so the flush-left block is in gofmt's "code block" canonical form
(blank line under `responses:` + TAB-indented children). That form scans
correctly because this framing strip composes with `yaml.RemoveIndent`'s
dedent/retab; the `*`-decorated block is gofmt-stable. A package doc comment
warns against re-spacing the tabs (gofmt would rewrite them).


<a id="a-1609"></a>

### #1609 — Vendor extension x-example for Dredd not working

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1609)

**✅ Works-as-designed / lock + 📖 (verified 2026-06-14).** x-* vendor
extensions are emitted on **both parameters and response headers** when declared
via an `Extensions:` block (mirroring swagger:meta's InfoExtensions /
Extensions), and both round-trip into the marshaled spec. The reporter's bare
`// x-example:` line is swallowed as the description — the supported form is an
`Extensions:` block. Locked by `fixtures/bugs/1609` + `TestCoverage_Bug1609` +
golden. 📖 Doc-site: Extensions: on params/headers; bare x-* lines become
description.


<a id="a-1613"></a>

### #1613 — Generate spec for response with string content type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1613)

**✅ Fixed / lock (verified 2026-06-14).** A `swagger:response` with a plain
`string` in:body field used to emit a response with no schema (description
only). It now emits `schema: {type: string}` (the primitive-body work, #2942).
Locked by `fixtures/bugs/1613` + `TestCoverage_Bug1613` + golden.


<a id="a-1635"></a>

### #1635 — Spec response generation without inner struct

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1635)

**✅ Fixed (`fix/go-swagger-1635`).** Root cause (general): in both
`responses.buildFromStruct` and `parameters.buildFromStruct`, an embed marked
`in: body` went through the `#2701` field-promotion recursion with `in: body`
inherited, so the embed's members were promoted individually. For responses each
promoted scalar overwrote `resp.Schema`, collapsing to `{type: string}` (pre-#2701
the fields landed under `headers`, the reporter's original complaint). For
parameters each member became a separate `in: body` parameter — invalid OAS2,
which allows at most one body parameter per operation.

Fix: a response/operation has a single body, so per-field promotion under
`in: body` is meaningless. When the embed's inherited `in:` is `body`, the embed
IS the body — built like a named `Body Foo` field (schema `$ref`s the embedded
struct). Responses route through a new `buildBodyEmbed`; parameters route the
embed through `processParamField` (the embedded field's name is its type name),
yielding one body parameter. Other `in:` values keep promoting fields (the
`#2701` header/query/path cases are untouched). Schemas have no `in:` concept, so
no analog exists there.

Cross-issue: this corrected the `#2701` embed-inheritance witness
(`embeddedBodyResponse`), whose `{type: string}` assertion had captured this very
bug — it now asserts `schema: {$ref: #/definitions/bodyPayload}` with no headers.
Only that one golden drifted (`embed_inheritance.json`); the parameters fix added
no golden drift.

Witnesses: `fixtures/bugs/1635` (`TestCoverage_Bug1635`: anonymous `in: body`
embed → single `$ref` body for both a `swagger:response` and a
`swagger:parameters` set, with named-`Body` guard rails) plus the updated
`embed-inheritance` fixture/test. Documented in `responses/README.md` and
`parameters/README.md` §in-discriminator / embedded-fields.


<a id="a-1665"></a>

### #1665 — can't combine a same annotation across multiple-lines

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1665)

**✅ Fixed / lock (verified 2026-06-14).** Multiple `swagger:parameters` lines on
one type are all honored — the shared param binds to operations listed across
every line (the foo op from line 1 and the bar op from line 2 both receive the
`id` path param). Locked by `fixtures/bugs/1665` + `TestCoverage_Bug1665` +
golden.


<a id="a-1670"></a>

### #1670 — Error operation (All): yaml: line 10: did not find expected key

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1670)

**✅ Assumed fixed (re-triaged 2026-06-17; no repro to assert).** The "yaml: did
not find expected key" error during generate spec is one of the various
Security-definition / YAML-body quirks. Those surfaces have since been reworked
(grammar2 + the Security real-YAML handling), so this class of failure is
believed resolved. No repro case is available, so this cannot be asserted
conclusively — recorded as assumed-fixed (Fred's call). The residual
"which annotation's YAML is at fault?" diagnostic-quality concern lives with
fail-loud diagnostics (forthcoming-features §8).


<a id="a-1708"></a>

### #1708 — Response annotation missing property type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1708)

**✅ Fixed / lock (verified 2026-06-14).** A body response whose fields are
pointers to other models no longer errors with "missing property type"
(follow-up of #619): the fields resolve to `$ref`s in an object schema. Locked
by `fixtures/bugs/1708` + `TestCoverage_Bug1708` + golden.


<a id="a-1713"></a>

### #1713 — How to label the example Response & Request

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1713)

**✅ Fixed / lock (verified 2026-06-14).** Response `examples:` keyed by mime
type (`examples: {application/json: {...}}`) are supported in the
`swagger:operation` YAML body. Locked by `fixtures/bugs/1713` +
`TestCoverage_Bug1713` + golden. (The struct-based `swagger:response`
example-by-mime form remains forthcoming feature §10 / #2871.)


<a id="a-1725"></a>

### #1725 — How to override version with generate spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1725)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** The spec version comes from
`swagger:meta` `Version:`. Build-time injection is a workflow concern, not a
scanner feature: drive codescan as a library and set `doc.Info.Version` on the
returned `*spec.Swagger` after `codescan.Run` (combine with `-ldflags -X` for
the build stamp), or post-process / overlay via `InputSpec`. **Doc-site action:**
in the version/meta docs, note that `Version:` is static and show the
library-side "set `Info.Version` after Run" recipe for build-time values.


<a id="a-1726"></a>

### #1726 — How to annotate lists?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1726)

**✅ Fixed (`fix/go-swagger-1726`).** Scope (per maintainer): NOT a full markdown
here-document (that is #3211) — just relax **list identification** so markdown
bullets (`* item`, `+ item`) are recognised exactly like the YAML dash form
(`- item`), uniformly across the annotation language.

Root cause: `trimContentPrefix` (`internal/parsers/grammar/preprocess.go`) did
`TrimLeft(s, " \t*/")`, eating a leading `* ` bullet as if it were godoc
decoration (dash was already preserved, for the `---` fence).

Fix (in the lexer pipeline, one place for wide impact): drop `*` from the
content-prefix strip set and **normalise** a leading `* `/`+ ` bullet to the
canonical `- ` in `trimContentPrefix`. To keep block-comment `* ` continuation
decoration from being mistaken for a bullet, `stripLine` now applies the
comment-kind raw-strip (`stripBlockContinuation` for `/* */`) BEFORE the
content-prefix trim. Because normalisation happens upstream on `Text`, every
downstream `- ` consumer handles markdown bullets with no per-site change:
prose descriptions, `Property.AsList` (consumes/produces/schemes/tags), enum
bodies. `Raw` is untouched, so YAML bodies stay strict YAML.

gofmt impact (checked): gofmt rewrites `*`/`+` doc-comment bullets to `-`
itself, so this only changes behaviour for non-gofmt'd source — and produces the
**same** result gofmt would. Verified by gofmt'ing the fixture and diffing
(every `*`/`+` → `-`, identical to the lexer). The fixture deliberately keeps
the `*`/`+` form to witness the scanner path end-to-end; `fixtures/` is excluded
from the gofmt formatters in `.golangci.yml`, matching existing non-gofmt-clean
bug fixtures.

Witnesses: `fixtures/bugs/1726` (Widget dash guard rail; Gadget `*`, Gizmo `+`
descriptions; a route with `*`-Produces / `+`-Consumes for the AsList path) +
`TestCoverage_Bug1726`; unit tests `TestNormalizeBullet` /
`TestTrimContentPrefixBullets`. No golden drift. Documented in
`grammar/README.md` §preprocess-contract + §property-shape.

**Doc-site action (📖):** document that list items may use `-`, `*`, or `+`
(all normalised to `-`), and that gofmt canonicalises them to `-`. Related:
#3211 (full markdown here-doc, not pursued here).


<a id="a-1727"></a>

### #1727 — Is there a way to automatically generate the swagger spec from the code?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1727)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** codescan **is** the
programmatic answer: `codescan.Run(*Options) (*spec.Swagger, error)` generates
the spec from source in-process — no CLI required. The library-usage
documentation now covers this. **Doc-site action:** ensure the "use codescan as
a library" page is linked from the getting-started/CLI docs so this
recurring question is answered up front.


<a id="a-1734"></a>

### #1734 — How to avoid name conflicts between service and dependencies

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1734)

**♻️ Duplicate of the name-identity track (#2637 / #2783) + 📖 (verified
2026-06-16).** Same-short-name structs across packages collapsing into one
`#/definitions/<name>` is exactly the collision the active name-identity /
cyclic-$ref work (#2637, #2783 — both 🛠) addresses; the auto-disambiguation it
needs is specified in forthcoming-features §14. Not wont-fix — being worked on.
**Doc-site action:** until it lands, document the explicit `swagger:model <name>`
disambiguation workaround.


<a id="a-1735"></a>

### #1735 — Golang 1.11 : "swagger generate spec" fail  > "assignment to entry in nil map"

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1735)

**✅ Resolved by engine rewrite (verified 2026-06-15).** The panic was in the
legacy `scan/route_params.go` (`setOpParams.Parse` → "assignment to entry in nil
map") while parsing a route's params. That entire `scan/` engine no longer
exists; grammar2 scans a route-with-params without panicking, witnessed by
`fixtures/bugs/1742` + `TestCoverage_Bug1742` (a `swagger:route` with a
`swagger:parameters` struct). Environmental/historical — no separate fixture.


<a id="a-1737"></a>

### #1737 — bson.ObjectId in swagger:response generates $ref

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1737)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** A field whose type resolves
to a `$ref` (an external named type such as `bson.ObjectId`) cannot carry a
sibling `description` in Swagger 2.0 — a `$ref` node has no room for one, so the
field-level description is dropped by default (matches the reporter's snippet 2).
The supported answer is the **`DescWithRef`** option: it rewrites the property to
`{description, allOf: [{$ref}]}`, preserving the description alongside the
reference (snippet 3's intent). Locked by `fixtures/bugs/1737` +
`TestCoverage_Bug1737` (asserts both the default drop and the `DescWithRef`
preservation) + golden. **Doc-site action:** document `DescWithRef` as the
answer for "my field's description disappears when the type becomes a `$ref`".


<a id="a-1742"></a>

### #1742 — Parameters (cross-package parameter structs)

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1742)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** A `swagger:parameters`
struct may live in a **different package** from the `swagger:route` /
`swagger:operation` that references it — as long as both packages are in the
scan set, the parameter is bound to the operation by its operation id. Locked by
`fixtures/bugs/1742` (param struct in subpackage `api`, route in the parent) +
`TestCoverage_Bug1742` + golden. **Doc-site action:** state that
`swagger:parameters`/`swagger:response` structs are collected across all scanned
packages and matched by operation id, so they need not sit beside the route.


<a id="a-1758"></a>

### #1758 — Newbie help: generate spec from source

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1758)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** The empty output
(`{"swagger":"2.0","paths":{}}`) means **no `swagger:*` annotations were found**
— the spec is built from annotations, not from the framework's (echo) route
registrations. Scanning multiple packages works (`./...` / a package list); the
reporter's code simply had no annotations. The "nothing produced, no error"
experience is the fail-loud gap tracked in `forthcoming-features.md` §8.2.
**Doc-site action:** a getting-started note that codescan reads annotations
(routes via `swagger:route`/`swagger:operation`), and that an empty spec means
none were found — not a scan-scope failure.


<a id="a-1761"></a>

### #1761 — Host url

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1761)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** The `host` field comes from
`swagger:meta` `Host:` and is static in the source. For per-environment values,
either **leave `Host:` empty** (consumers then resolve against the serving host
— the usual choice for env-portable specs) or drive codescan as a library and
set `doc.Host` on the returned `*spec.Swagger` after `Run`. A build/serve-time
concern, like #1725. **Doc-site action:** document the empty-host /
set-`Host`-programmatically recipe for multi-environment APIs.


<a id="a-1772"></a>

### #1772 — Post Example

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1772)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** A POST request body is
documented with a `swagger:parameters` struct whose field is marked `in: body`
(the field's type becomes the body schema). **Doc-site action:** a worked
"body parameter for a POST" example in the parameters docs.


<a id="a-1795"></a>

### #1795 — Is it possible to generate a header with Basic base64(data) for a post request?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1795)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** A `Authorization: Basic
<base64>` header is **HTTP Basic auth**, a first-class OpenAPI 2.0 security
scheme — not a manually-declared header parameter. Declare it once in
`swagger:meta` `SecurityDefinitions:` as `{type: basic}` and reference it via a
`Security:` requirement (on `swagger:meta` for global, or per
`swagger:route`/`swagger:operation`). The base64 encoding is the transport
detail of Basic auth and is not modelled in the spec. **Doc-site action:**
document Basic (and the other OAS2 security schemes) under a security how-to,
steering users away from hand-rolling an `Authorization` header param.


<a id="a-1815"></a>

### #1815 — Is this possible to create/regenerate spec from swagger generated server files?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1815)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Yes — go-swagger-generated
server code carries `swagger:*` annotations (operations, models, parameters,
responses), so pointing codescan at those sources regenerates the spec. The
round-trip is annotation-driven like any other scan. **Doc-site action:** note
in the usage docs that generated (or hand-annotated) server code is a valid
scan input, so spec↔server round-tripping works.


<a id="a-1828"></a>

### #1828 — generate/spec: route responses tag description

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1828)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** The reporter's inline
`200: description:Message Success` form (snippet 1) now produces the valid
`responses: {200: {description: "Message Success"}}` (snippet 2, the desired
output); the legacy engine emitted the invalid `$ref: '#/responses/'` with an
empty name (snippet 3). Locked by `fixtures/bugs/1828` + `TestCoverage_Bug1828`
(asserts the inline description and the absence of a spurious `$ref`) + golden.


<a id="a-1852"></a>

### #1852 — Schema error for delete operation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1852)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** A 204 response whose
`swagger:response` type has no doc comment now still emits the OAS2-required
`description` key (empty string), so editor.swagger.io no longer reports
"missingProperty: description". The legacy engine omitted the key entirely.
Locked by `fixtures/bugs/1852` + `TestCoverage_Bug1852` (asserts the key is
present even with no doc comment) + golden. **Doc-site action:** note that a
response's description comes from its type's godoc — give response types a doc
comment so the emitted description is meaningful rather than empty.


<a id="a-1865"></a>

### #1865 — swagger:parameters split into new lines

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1865)

**♻️ Duplicate of #1665 + 📖 (verified 2026-06-15).** A long `swagger:parameters`
operation-id list is split by writing **multiple `swagger:parameters` lines** on
the struct — each line binds it to more operations, and they accumulate. This is
exactly what #1665 (`fixtures/bugs/1665` + `TestCoverage_Bug1665`) already locks.
**Doc-site action:** document the multi-line `swagger:parameters` technique for
sharing one parameter struct across many operations.


<a id="a-1867"></a>

### #1867 — PATCH operation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1867)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** PATCH works through both
annotation styles: `swagger:route PATCH ...` emits a `patch` operation and
accepts a `swagger:parameters` body struct, and `swagger:operation PATCH ...`
emits its `patch` operation from the YAML body. The reporter's "swagger:operation
silently fails / swagger:route can't specify the body" is resolved. Locked by
`fixtures/bugs/1867` + `TestCoverage_Bug1867` (both forms) + golden.


<a id="a-1881"></a>

### #1881 — Key words in comments

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1881)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** The array-of-object response
the reporter struggled with now produces a valid schema: a `swagger:response`
whose body field is a slice of a model emits `{type: array, items: {$ref}}`.
Locked by `fixtures/bugs/1881` + `TestCoverage_Bug1881` + golden. The other half
of the issue — "I can't find the full list of annotations/keywords" — is purely
a documentation gap. **Doc-site action:** a complete annotation/keyword
reference (the migration of `docs/*.md` already on the doc-site roadmap), and an
array-response example.


<a id="a-1887"></a>

### #1887 — generate spec: file type support.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1887)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Swagger 2.0 `type: file`
(formData only) is supported via the **`swagger:file`** marker on a
`swagger:parameters` field marked `in: formData` — it emits a `{in: formData,
type: file}` parameter. Locked by `fixtures/bugs/1887` + `TestCoverage_Bug1887`
+ golden. **Doc-site action:** document the `swagger:file` marker and the
`consumes: multipart/form-data` pairing for file uploads.


<a id="a-1891"></a>

### #1891 — go-swagger doesn't work for group type ?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1891)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** Both halves work now: a
`swagger:model` declared inside a grouped `type ( ... )` block is discovered, and
a `swagger:route` declared **inside a function body** is parsed (params bound,
body `$ref` resolved). The reporter's snippets reflect the older engine's
sensitivity to declaration grouping and single- vs multi-file layout; current
codescan is layout-insensitive here. Locked by `fixtures/bugs/1891` +
`TestCoverage_Bug1891` (grouped-type model + func-body route) + golden. (The
reporter's leftover confusion was using the same identifier — `FooInput` — as
both the route operation id and the params struct; unrelated to grouping.)


<a id="a-1913"></a>

### #1913 — Sub-types not generated from go discriminated type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1913)

**✅ Fixed by grammar2 / lock + 📖 (verified 2026-06-15).** Interface-based
discriminator models work: a `swagger:model` interface with a `discriminator:
true` member emits a base definition carrying `discriminator`, and the struct
subtypes that embed it via `swagger:allOf` emit `allOf: [{$ref base}, {own
props}]`. Locked by `fixtures/bugs/1913` + `TestCoverage_Bug1913` (asserts the
base discriminator and both subtypes' allOf) + golden. The reporter's residual
concern — subtypes are only emitted under `-m`, which over-generates — is a
**discovery** gap, now registered as forthcoming-features §15 (auto-include a
referenced discriminated base's allOf subtypes without `-m`; refines §12).
**Doc-site action:** keep the discriminated-types how-to, and note the current
`-m` requirement for subtype emission until §15 lands.


<a id="a-1925"></a>

### #1925 — parameters do not support interface

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1925)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** A body parameter typed
`[]map[string]interface{}` (the reporter's `Report` field) no longer aborts
("unsupported for a schema"); it produces a valid `{type: array, items: {type:
object, additionalProperties: {}}}`. Locked by `fixtures/bugs/1925` +
`TestCoverage_Bug1925` + golden. (The reporter's `expr (XXXX:...)` abort was
ultimately an unresolved external import, an orthogonal load issue.)


<a id="a-1931"></a>

### #1931 — how to generate spec for package

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1931)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** To cover all sub-packages
rather than just `doc.go`, scan with `./...` (or an explicit package list) —
codescan loads the whole package graph in scope and collects annotations from
every package. The empty `paths` in the reporter's output reflects scanning only
the meta package. **Doc-site action:** a getting-started note on scan scope
(`./...` vs a single file/package), paired with #1758's "annotation-driven,
empty means none found" guidance.


<a id="a-1934"></a>

### #1934 — model declared in function not picked up

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1934)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** Annotated types declared
**inside a function body** are now discovered: the `swagger:model someResponse`
local type is emitted and resolves the route's `default: body:someResponse`
`$ref`, and the `swagger:parameters` local type contributes the query parameter.
Locked by `fixtures/bugs/1934` + `TestCoverage_Bug1934` + golden.


<a id="a-1955"></a>

### #1955 — swagger:parameters and swagger:operation in different go package

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1955)

**✅ Fixed by grammar2 / lock + 📖 (verified 2026-06-15).** A `swagger:operation`
binds a `swagger:parameters` struct defined in a **different package** by
operation id — the cross-package collection works for `swagger:operation` just
as #1742 showed for `swagger:route`. Locked by `fixtures/bugs/1955` (operation in
the parent package, params struct in subpackage `paramsx`) +
`TestCoverage_Bug1955` + golden. **Doc-site action:** the same "parameters are
collected across all scanned packages, matched by operation id" note covers both
`swagger:route` (#1742) and `swagger:operation` (this issue).


<a id="a-1958"></a>

### #1958 — Ability to skip embedded tags

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1958)

**✅ Fixed by grammar2 / lock + 📖 (verified 2026-06-15).** A vendor extension in
the `swagger:operation` YAML body is preserved **in full**, including nested keys
that share a name with a known keyword: the `responses:` map *inside*
`x-amazon-apigateway-integration` survives verbatim and is not confused with the
operation's own responses. Locked by `fixtures/bugs/1958` + `TestCoverage_Bug1958`
+ golden. **Doc-site action:** document that arbitrary `x-*` extensions on an
operation are passed through untouched, nested structure included.

> ✅ **Adjacent finding FIXED & MERGED (2026-06-15, merge `01b817e`) — was branch `fix/gofmt-operation-yaml-tabs`.**
> A gofmt'd swagger:operation YAML body uses **tab** indentation (invalid YAML),
> which silently dropped responses/extensions — the operations counterpart of
> quirk F7 (which covered swagger:meta only). Root: `yaml.RemoveIndent` stripped
> the prose-key width off the children's lone leading tab and flattened the
> nesting. Fix: expand leading tabs→spaces before the strip. Witness
> `fixtures/quirks/gofmt-operation` + `TestQuirk_GofmtOperationYAML`; also
> regenerated `bugs_3138_schema.json` (that fixture was itself tab-form and had
> silently dropped its `200` response). Discovered while verifying #1867/#1955/#1958.


<a id="a-1974"></a>

### #1974 — Unable to run `swagger.go`

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1974)

**✅ Resolved by toolchain / N-A (verified 2026-06-15).** The error is an
ancient vendored `golang.org/x/tools` compile failure (`obj.IsAlias undefined`,
`types.SizesFor` undefined) — a stale go-swagger build against a mismatched Go
version, not a codescan scan issue. N/A under the modern module-aware
`go/packages` loader. **Doc-site action:** none beyond a general "keep the
toolchain current" note.


<a id="a-2002"></a>

### #2002 — Generate spec fails with invalid type error

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2002)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** A swagger:response body
field whose type lives in **another package** resolves to a `$ref` definition.
The reporter's `unsupported type "invalid type"` came from the old
`GO111MODULE=off` go/loader (vendored, module-blind); the module-aware
`go/packages` loader resolves the cross-package type. Locked by
`fixtures/bugs/2002` (body type in subpackage `api`) + `TestCoverage_Bug2002` +
golden.


<a id="a-2013"></a>

### #2013 — swagger generate spec panic: runtime error: index out of range

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2013)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** The old engine panicked
(`index out of range` in `schemaValidations.SetEnum`, during `buildEmbedded`)
while parsing an enum on a promoted field. grammar2 parses an enum on an
embedded field without panicking and carries the values onto the composed model.
Locked by `fixtures/bugs/2013` (Base.Kind enum promoted into Thing) +
`TestCoverage_Bug2013` + golden.


<a id="a-2020"></a>

### #2020 — Parameters not detected with multiple structs in one statement

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2020)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** Multiple
`swagger:parameters` structs declared in a single grouped `type ( ... )` block
are each detected and bound to their operation by id (the reporter's "params not
detected with multiple structs in one statement / annotation must be above
type"). Locked by `fixtures/bugs/2020` (two param structs in one group →
alphaOp/betaOp) + `TestCoverage_Bug2020` + golden. **Known narrow edge (not this
issue's core):** an anonymous *inline* body struct field marked `// in:body` +
`// swagger:name -` is still dropped to a query parameter; the grouped-declaration
detection is independent of it.


<a id="a-2027"></a>

### #2027 — Embedded struct Support

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2027)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Embedded (anonymous) struct
support is implemented: an embedded struct's fields are promoted into the
composed model, witnessed across #413 (the original embedded-struct error, fixed),
#2013 (enum on a promoted field) and #2038. The long-open question (refers to
#413) is resolved. **Doc-site action:** document embedded-struct composition
(promotion of anonymous fields; allOf where applicable). Note the json-tagged
embed nuance tracked separately as #2038.


<a id="a-2038"></a>

### #2038 — Swagger ignore json tags for embedded structures

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2038)

**✅ Fixed / lock (verified 2026-06-15; merged `1da0aeb`).** A json-tagged
embedded struct now NESTS under its tag instead of promoting its fields, matching
encoding/json. The previously-RED `fixtures/bugs/2038` + `TestCoverage_Bug2038`
(tagged embed → nested property; untagged embed → promoted) is green on
`fix/backlog-lot1`.


<a id="a-2062"></a>

### #2062 — Cannot add security and SecurityDefinitions in swagger:operation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2062)

**✅ Fixed / lock (verified 2026-06-15).** A `security:` requirement in the
swagger:operation YAML body is emitted on the operation. `SecurityDefinitions` is
a global `swagger:meta` concept in OpenAPI 2.0 (not a per-operation field), so
that half of the reporter's ask is N/A by spec. Locked by `fixtures/bugs/2062` +
`TestCoverage_Bug2062` + golden.


<a id="a-2064"></a>

### #2064 — add example to string parameter in request body

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2064)

**✅ Fixed / lock (verified 2026-06-15).** A body parameter's `example` and
`default` are now emitted (previously missing) — carried on the parameter object
alongside the body `schema`. Locked by `fixtures/bugs/2064` +
`TestCoverage_Bug2064` + golden. (Minor placement nuance: for a body parameter
these sit on the parameter object rather than inside `schema`; the values are
present, which resolves the "missing" report.)


<a id="a-2106"></a>

### #2106 — `swagger generate spec` ignores `Extensions` on models when type is not an array

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2106)

**✅ Fixed / lock (verified 2026-06-15).** A field-level `Extensions:` block emits
its `x-*` vendor extensions on BOTH scalar and array fields (the reporter saw
them only on array fields). Locked by `fixtures/bugs/2106` (scalar `name` +
array `friends`, each with an extension) + `TestCoverage_Bug2106` + golden.


<a id="a-2119"></a>

### #2119 — Add flag to skip generation of `x-go-name`

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2119)

**✅ Implemented + 📖 (verified 2026-06-15).** The `SkipExtensions` option
suppresses the scanner-derived `x-go-name` / `x-go-package` vendor extensions —
exactly the requested flag (raised because those clash across same-named types,
cf. #1734 §14). Locked by `fixtures/bugs/2119` + `TestCoverage_Bug2119` (asserts
present by default, absent under `SkipExtensions`) + golden. **Doc-site action:**
document `SkipExtensions` in the Options reference.


<a id="a-2125"></a>

### #2125 — Parsing meta info comments can parse fields wrong.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2125)

**✅ Fixed by grammar2 / lock (verified 2026-06-16).** Markdown content in a
`swagger:meta` block (headings, prose mentioning "Api-Version", code fences)
stays in the info `description` and does NOT derail field parsing — the real
`Version:` field is read correctly (the reporter's "Api-Version read as Version,
ending comment parsing" no longer reproduces). Locked by `fixtures/bugs/2125` +
`TestCoverage_Bug2125` + golden.


<a id="a-2126"></a>

### #2126 — Reference swagger models under a specific package

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2126)

**♻️ Duplicate of the name-identity track (#2637 / #2783) + 📖 (verified
2026-06-16).** Referencing same-named models from different packages (without
mixing their definitions/examples) is the same definition-name collision being
fixed on the active name-identity / cyclic-$ref track (#2637, #2783); the
disambiguation design is forthcoming-features §14 (origin #1734). **Doc-site
action:** document the `swagger:model <name>` workaround meanwhile.


<a id="a-2127"></a>

### #2127 — How to swagger ignore specific lines in the documentation?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2127)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A single line / field is
omitted by putting `swagger:ignore` on that field (locked for fields by #2311).
**Doc-site action:** document that `swagger:ignore` drops the annotated field
(and applies to a single struct field, not just whole types).


<a id="a-2133"></a>

### #2133 — spec generating schema and type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2133)

**✅ Fixed / lock (verified 2026-06-16).** A path parameter is emitted as a
simple `{type:string, in:path}` with NO sibling `schema` — the double-emission
(both `schema` and `type`) that failed validation is gone. Locked by
`fixtures/bugs/2133` + `TestCoverage_Bug2133` + golden.


<a id="a-2160"></a>

### #2160 — Example of array of structs is 0 valued when generating spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2160)

**✅ Fixed / lock + 📖 (verified 2026-06-16).** An array-of-structs field accepts
an inline JSON-array `example:` and emits it as an array of objects (not
zero-valued). Locked by `fixtures/bugs/2160` + `TestCoverage_Bug2160` + golden.
The reporter's **multi-line YAML-list** example syntax is instead kept as a raw
string — the example-coercion gap tracked in forthcoming-features §2.1.
**Doc-site action:** document the inline-JSON example form for arrays.


<a id="a-2172"></a>

### #2172 — property comment does not get generated into swagger.yaml

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2172)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A property comment on a field
whose type resolves to a `$ref` is dropped by default (a `$ref` node cannot carry
a sibling `description` in Swagger 2.0), but `DescWithRef` wraps it in `allOf` and
keeps the comment. Identical mechanism to #1737. **Doc-site action:** document
`DescWithRef` for keeping descriptions on `$ref`-typed fields.


<a id="a-2183"></a>

### #2183 — Different spec produced for same codebase

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2183)

**✅ Works-as-designed (verified 2026-06-16).** The generated spec is
deterministic, including for a swagger-annotated type that embeds another: 4
repeated scans of an embed-bearing model produce byte-identical output (keys in
sorted order). The map-ordering nondeterminism the reporter saw no longer
reproduces; covered by #2299 (deterministic golden harness, #2762).


<a id="a-2184"></a>

### #2184 — spec generation fails to replace $ref with indicated type when using swagger:type annotation on struct

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2184)

**✅ Fixed / lock (verified 2026-06-16).** A struct type carrying `swagger:type
int64`, used as a parameter field, makes the parameter an `integer` (format
int64) — not a `$ref` to the struct. The v0.17 regression is gone. Locked by
`fixtures/bugs/2184` + `TestCoverage_Bug2184` + golden.


<a id="a-2208"></a>

### #2208 — Scanner tests are excluded from build ?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2208)

**✅ Fixed / no longer applicable (verified 2026-06-16).** The `// +build !go1.11`
exclusion dates to a Go version retired years ago; codescan requires modern Go
(go1.11+ modules) so the tag is moot — no scanner tests are gated by it. Resolved
by default, no change needed.


<a id="a-2210"></a>

### #2210 — unknown field 'URL' in struct literal of type spec.ContactInfo

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2210)

**✅ Resolved by toolchain / N-A (verified 2026-06-16).** The build error
("unknown field URL/Name/Email in spec.ContactInfo/License") is a
`go-openapi/spec` API version skew in an old build, not a codescan scan issue.
N/A with current pinned dependencies.


<a id="a-2218"></a>

### #2218 — Unable to connect a go-swagger parameter to a route

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2218)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A `swagger:parameters`
struct binds to its route when the struct's operation id matches the route's
operation id — the query parameter then appears on the operation. The reporter's
param simply had a mismatched id. Locked by `fixtures/bugs/2218` +
`TestCoverage_Bug2218` + golden. **Doc-site action:** stress that the
`swagger:parameters <opid>` must match the route/operation id exactly.


<a id="a-2228"></a>

### #2228 — how to give empty summary in swagger:route

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2228)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** The auto-derived summary is
the first prose line of the route comment; **omit that line** and the operation
has no `summary` (a route with only `responses:` emits none). Locked by
`fixtures/bugs/2228` + `TestCoverage_Bug2228` + golden. **Doc-site action:**
document that the first prose line becomes the summary and how to omit it.


<a id="a-2230"></a>

### #2230 — [Question] How to define an example in json.RawMessage field of a struct

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2230)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A `json.RawMessage` field
renders as an **open (typeless) schema** — i.e. "any JSON", which matches
RawMessage's meaning — while typed sibling fields keep their `example`. The raw
bytes are not forced to an int array. To constrain or example the RawMessage,
annotate it (`swagger:type` / `swagger:strfmt`). Locked by `fixtures/bugs/2230` +
`TestCoverage_Bug2230` + golden. **Doc-site action:** document that
json.RawMessage → open schema and how to type/example it.


<a id="a-2232"></a>

### #2232 — Unable to generate tags with spaces using the spec generation tool

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2232)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** A multi-word tag is expressed
via the **swagger:operation** YAML body `tags:` list (`- Thing Management`). The
**swagger:route** line cannot carry spaced tags — its tags are space-delimited
tokens. Locked by `fixtures/bugs/2232` + `TestCoverage_Bug2232` + golden.
**Doc-site action:** document that spaced/multi-word tags require the
swagger:operation YAML form.


<a id="a-2233"></a>

### #2233 — Generate spec: Unable to find responses defined in other package using swagger:operation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2233)

**✅ Fixed / lock (verified 2026-06-16).** A swagger:operation YAML body that
`$ref`s a `swagger:response` defined in ANOTHER package resolves — the response
and its body model are emitted (no "$refs must reference a valid location").
Locked by `fixtures/bugs/2233` (response in subpackage `model`) +
`TestCoverage_Bug2233` + golden.


<a id="a-2245"></a>

### #2245 — how to write swagger:response schema that produces application/xml

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2245)

**✅ Works-as-designed + 📖 (verified 2026-06-16).** The produced media type is
set with `produces:` (in the swagger:operation YAML body, or `Produces:` on
swagger:route/swagger:meta) — `produces: [application/xml]` yields an XML
operation. A response object itself does not carry a mime type in OpenAPI 2.0.
Locked by `fixtures/bugs/2245` + `TestCoverage_Bug2245` + golden. **Doc-site
action:** document `produces` for non-JSON media types.


<a id="a-2248"></a>

### #2248 — Dealing with time.Duration in response's header

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2248)

**✅ Fixed / lock (verified 2026-06-16).** A response header typed `time.Duration`
resolves to `{type: integer, format: int64}` — it carries a type, so the spec is
valid. The reporter's "headers.<name>.type in body is required" no longer
reproduces. Locked by `fixtures/bugs/2248` + `TestCoverage_Bug2248` + golden.


<a id="a-2251"></a>

### #2251 — Problems getting map with non-string keys serialized in spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2251)

**✅ Fixed / lock + 📖 (landed 2026-06-17, commit `f354ff0`).** `buildFromMap`
used to emit `additionalProperties` only for `string` / `TextMarshaler` keys and
silently dropped every other key to a typeless property. It now also emits
`{type:object, additionalProperties:V}` for **integer-kind keys** (`map[int]V`,
`map[int64]V`), which `encoding/json` marshals as string keys; a genuinely
unrepresentable key (e.g. `map[float64]V`) raises a `CodeUnsupportedType` warning
instead of silently emitting an open schema — the fail-loud half of
forthcoming-features §18. Locked by `fixtures/bugs/2251` + `TestCoverage_Bug2251`.
📖 Doc-site: document which key types are supported (string, integer,
TextMarshaler).


<a id="a-2286"></a>

### #2286 — Model accepted as Response

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2286)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Naming a `swagger:model`
directly as a route response (`200: logoutResponse`) is accepted and yields a
valid response with a `$ref` to that model — the lenient shorthand for `200:
body:logoutResponse`. Locked by `fixtures/bugs/2286` + `TestCoverage_Bug2286` +
golden. **Doc-site action:** document `NNN: <model>` vs `NNN: body:<model>`
(both produce a $ref body response).


<a id="a-2294"></a>

### #2294 — Unable to generate Swagger spec with more than one security header using "AND" logic.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2294)

**✅ Fixed / lock (verified 2026-06-16; merged `e64c662`).** The meta `Security:`
block is now parsed as real YAML, so two schemes in one requirement (`- x: []`
+ indented `y: []`) form a single AND requirement (no longer split into two OR
requirements). `fixtures/bugs/2294` + `TestCoverage_Bug2294` green on
`fix/backlog-lot1`.


<a id="a-2296"></a>

### #2296 — panic,embedded meet anonymous

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2296)

**✅ Fixed / lock (verified 2026-06-15).** Embedding a struct that itself has an
anonymous-struct field no longer panics during schema build; the promoted
anonymous-struct property is emitted as a nested object. Locked by
`fixtures/bugs/2296` + `TestCoverage_Bug2296` + golden.


<a id="a-2299"></a>

### #2299 — Generated swagger schema is not deterministic

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2299)

**✅ Works-as-designed (verified 2026-06-15).** The generated schema is
deterministic: 5 repeated scans of a many-field model produce byte-identical
output (definitions and properties are emitted in sorted key order by
go-openapi/spec marshaling). The nondeterminism the reporter saw no longer
reproduces; the deterministic golden harness (#2762) guards it.


<a id="a-2305"></a>

### #2305 — Query params and path params, enum dropdown

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2305)

**✅ Fixed / lock (verified 2026-06-15).** A query parameter with an enum is
clean and valid — the `enum` is present, with no illegal `schema` on the
parameter and no duplicated description. Locked by `fixtures/bugs/2305` +
`TestCoverage_Bug2305` + golden.


<a id="a-2311"></a>

### #2311 — swagger:ignore documentation is incomplete with respect to fields

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2311)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** `swagger:ignore` on a struct
FIELD drops that property from the model (the behaviour added in #1497). The
issue is purely the doc gap — the reference only mentioned types/whole
declarations. Locked by `fixtures/bugs/2311` + `TestCoverage_Bug2311` + golden.
**Doc-site action:** document that `swagger:ignore` also applies to individual
struct fields.


<a id="a-2317"></a>

### #2317 — Extension x-nullable on a pointer has no effect

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2317)

**✅ Fixed / lock (verified 2026-06-15).** An `x-nullable: true` Extensions block
on a **pointer** field is applied: the field becomes `allOf: [{$ref}]` carrying
`x-nullable: true`. Locked by `fixtures/bugs/2317` + `TestCoverage_Bug2317` +
golden.


<a id="a-2353"></a>

### #2353 — Generation of valid spec file with body and request param

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2353)

**✅ Fixed / lock (verified 2026-06-15).** An operation mixing a path parameter
and a body parameter produces a valid parameter set — the body param is a clean
`{in:body, schema:{$ref}}` with no forbidden sibling `type`, the path param is
`{in:path, type:string}`. Locked by `fixtures/bugs/2353` + `TestCoverage_Bug2353`
+ golden.


<a id="a-2371"></a>

### #2371 — Generate spec fails for response with array of objects

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2371)

**✅ Works-as-designed (verified 2026-06-15).** A response that is an array of
objects produces a valid `{type:array, items:{$ref}}` schema — covered by #1881
(`fixtures/bugs/1881` + `TestCoverage_Bug1881`). The "fails when a response is an
array of objects" report no longer reproduces.


<a id="a-2379"></a>

### #2379 — unsupported type "invalid type" error when using Linux binary

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2379)

**✅ Works-as-designed (verified 2026-06-15).** Referencing a schema in another
package resolves to a `$ref` (Outer.inner → `$ref` Inner from the subpackage).
The "unsupported type 'invalid type'" from the old v0.25 binary was the
GO111MODULE/go-loader era; the module-aware `go/packages` loader resolves it
(covered by #2002 / #2520 / #2549).


<a id="a-2383"></a>

### #2383 — Using inline struct inside of function that returns http.HandlerFunc can't generate spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2383)

**✅ Works-as-designed (verified 2026-06-15).** An annotated struct declared
inside a function — including one that returns `http.HandlerFunc` — is discovered
and emitted. Covered by #1934 (`fixtures/bugs/1934` + `TestCoverage_Bug1934`); the
"unable to find package and source file" report no longer reproduces.


<a id="a-2384"></a>

### #2384 — pattern containing '\n' is interpreted when generating comment producing illegal output.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2384)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** codescan reads a `pattern:`
value VERBATIM — `^[^ \t\n](.*[^ \t\n])*$` is emitted with the backslash
escapes preserved as the two characters they are (`\t`, `\n`), not turned into
literal tab/newline. The reporter's broken multi-line output came from
go-swagger's **spec→code** generator (it interpreted `\n` when writing the
comment), not from codescan's code→spec scan. Locked by `fixtures/bugs/2384` +
`TestCoverage_Bug2384` + golden. **Doc-site action:** none for codescan; the
codegen-side escaping belongs to go-swagger.


<a id="a-2396"></a>

### #2396 — model enum recognizion and handling spaces

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2396)

**✅ Fixed / lock (verified 2026-06-16; merged `8c3b970`).** The bracketed
`enum: [a, b, c]` form now strips the surrounding `[`/`]`, yielding the same enum
as the unbracketed form (`["issues","pulls","projects"]`). `fixtures/bugs/2396` +
`TestCoverage_Bug2396` green on `fix/backlog-lot1`.


<a id="a-2398"></a>

### #2398 — Warn about duplicate definitions

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2398)

**♻️ Duplicate of the name-identity track (#2637 / #2783) + 📖 (verified
2026-06-16).** Same-id definitions silently overriding each other is the
collision the active name-identity / cyclic-$ref work (#2637, #2783) tackles; the
warn-on-overwrite diagnostic belongs to that effort (cf. §8 fail-loud, §14).
**Doc-site action:** document that definition names must be unique +
`swagger:model <name>` disambiguation.


<a id="a-2403"></a>

### #2403 — go swagger security auth0

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2403)

**✅ Fixed / lock (verified 2026-06-16; merged `048d607`).** The meta `Security:`
sequence now parses correctly: `- auth0: []` yields a requirement keyed by
`auth0` with empty scopes (the `- ` sequence marker is no longer glued onto the
key). `fixtures/bugs/2403` + `TestCoverage_Bug2403` green on `fix/backlog-lot1`.
(Sibling meta-Security gaps #2294 (AND multi-key) and #2479 (route opt-out)
remain open — that fix did not cover them.)


<a id="a-2407"></a>

### #2407 — how add example to yml from golang code

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2407)

**✅ Fixed / lock (verified 2026-06-15).** An `example: [1, 2, 3]` on an
array-typed body response is emitted on the response schema (previously absent).
Locked by `fixtures/bugs/2407` + `TestCoverage_Bug2407` + golden.


<a id="a-2409"></a>

### #2409 — Annotate structures with extensions

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2409)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** A TYPE-level `Extensions:`
block emits its `x-*` extensions at the definition level. Locked by
`fixtures/bugs/2409` + `TestCoverage_Bug2409` + golden. The reporter's specific
`x-go-type` IMPORT-directive semantics is the separate design question parked as
poison-queue #2924 — but the extension *mechanism* works for any `x-*` key.
**Doc-site action:** document type-level Extensions blocks.


<a id="a-2412"></a>

### #2412 — Cannot generate a parameter with type "file"

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2412)

**✅ Works-as-designed + 📖 / ♻️ (verified 2026-06-15).** The reporter used a
non-existent Go type `file` (hence "unsupported type 'invalid type'"). The
supported mechanism is the `swagger:file` marker on a formData field — already
locked by **#1887** (`fixtures/bugs/1887` + `TestCoverage_Bug1887`). Duplicate of
#1887. **Doc-site action:** document `swagger:file` for file-upload parameters.


<a id="a-2417"></a>

### #2417 — Embedding of aliased type

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2417)

**✅ Fixed (2026-06-14).** Embedding an *anonymous* defined type whose
underlying struct lives in a **different package** than the type
(`a.AnotherPackageAlias`, underlying `color.Color`, embedded from package `b`)
promoted no fields — the model came out a bare empty object. The other three
matrix cells already worked: embedding a plain cross-package type
(`color.Color`) and a same-package defined type (`color.SamePackageAlias`) both
promote `hue`; a named cross-package field gets a `$ref`.

Root cause: `tpe.Underlying()` collapses the defined-type chain straight to the
`*types.Struct`, so `buildFromStruct` ran with the embedding type's `decl`
(package `a`), but the promoted fields' source lives in `color.Color`'s file.
`structFieldCarrier` resolved each field's AST with
`FindASTField(decl.File, fld.Pos())`, which returns nil when the field isn't in
that file — so every cross-source-file field was silently dropped. Fix: a
fallback to the new `ScanCtx.FileForPos(pkgPath, pos)`, which locates the
field's own source file via the shared FileSet; the carrier reads its json tag
and doc from there. The fallback only fires when the primary lookup misses, so
the common single-file path is unchanged (zero golden churn).

The same root cause also affected a same-package **cross-file** case: a
transparent alias to a struct defined in a sibling file. `responses_test.go`'s
`TestParseResponses_TransparentAliases` previously captured the buggy empty
`payload` object (contradicting its own "expand inline and retain field
metadata" intent); its golden now carries the promoted `id`/`name` fields, and
explicit `payload.Properties` assertions were added.

Verified by `TestCoverage_Bug2417` (now green; three guard-rail cells +
`CrossAliasEmbed`) and the strengthened transparent-alias test. Branch
`fix/go-swagger-2417`; fixtures `fixtures/bugs/2417/{color,a,b}`. Documented in
`internal/builders/schema/README.md` §embedded.


<a id="a-2419"></a>

### #2419 — feature: support custom swagger type for struct field

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2419)

**✅ Fixed / lock (verified 2026-06-15).** A `swagger:type string` on a field
whose type is an external/imported struct (e.g. a protobuf `wrappers.StringValue`)
overrides it to a plain string instead of a `$ref` to the external type. Locked by
`fixtures/bugs/2419` (field of a subpackage type) + `TestCoverage_Bug2419` +
golden.


<a id="a-2441"></a>

### #2441 — File upload how to describe in annotations?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2441)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** A raw (non-multipart) binary
body is described with a string field marked `swagger:strfmt binary` → `{type:
string, format: binary}`. OpenAPI 2.0 `type: file` is **formData-only** (cf.
#1887/#2412), so a raw request body uses a binary-format string, not a file.
Locked by `fixtures/bugs/2441` + `TestCoverage_Bug2441` + golden. (A `[]byte`
field maps to the literal Go shape — an array of uint8; use a string +
`swagger:strfmt binary` for a binary body.) **Doc-site action:** document the
formData-file vs binary-body distinction.


<a id="a-2479"></a>

### #2479 — How to disable security on a route but keep on all endpoints?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2479)

**✅ Fixed / lock (verified 2026-06-16; merged `b4f530b`).** A route with
`Security: []` now emits an explicit empty `security: []` on the operation —
overriding global security per OAS2 (no longer dropped/inherited).
`fixtures/bugs/2479` + `TestCoverage_Bug2479` green on `fix/backlog-lot1`.


<a id="a-2483"></a>

### #2483 — Generation from code not working with allOf

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2483)

**✅ Fixed / lock (verified 2026-06-14).** No duplication. `Pet` emits exactly
two allOf arms — `{$ref: SimpleOne}` and a single inline merge object carrying
the promoted `did`/`cat` (from the unannotated `Something` embed) plus `notes`
and `extra`. No repeated arms, no "circular ancestry" validation error. Locked
by `fixtures/bugs/2483` + `TestCoverage_Bug2483` + golden `bugs_2483_schema.json`.

<a id="a-2520"></a>

### #2520 — Generate spec fails with unsupported type "invalid type"

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2520)

**✅ Resolved / lock (verified 2026-06-15).** A struct with a custom MarshalJSON
and only unexported fields no longer aborts with "unsupported type 'invalid
type'" (that came from an unresolved import, like #2002); it emits an object and
the using field gets a `$ref`. codescan works at the type level and cannot infer
the custom-marshaled wire shape — annotate the wrapper with `swagger:type` /
`swagger:strfmt` to declare it. Locked by `fixtures/bugs/2520` +
`TestCoverage_Bug2520` + golden. **Doc-site action:** document the
swagger:type/strfmt override for custom-marshaled types.


<a id="a-2528"></a>

### #2528 — go-swagger documentation is not updated for swagger enum

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2528)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Pure documentation gap:
`swagger:enum` works (extensively covered by `fixtures/enhancements/enum-*` and
the #2013 embedded-enum lock) but the usage is under-documented. **Doc-site
action:** a `swagger:enum` reference section — named-type const collection, the
inline `enum:` keyword, and the `x-go-enum-desc` extension.


<a id="a-2539"></a>

### #2539 — Generate spec with additionalProperties

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2539)

**✅ Fixed via feature §17 (landed 2026-06-17, `feat/additional-properties`
Phase 2).** Originally there was no annotation to set `additionalProperties`
explicitly on a model — `additional properties: false` was silently absorbed
into the description. The additionalProperties/patternProperties feature now
delivers full control: the `swagger:additionalProperties <spec>` type/model
marker, the `additionalProperties: <spec>` field keyword (`true | false |
TypeSpec`), and the typed `swagger:patternProperties "<re>": <spec>, …` marker
(plus the SimpleSchema safeguard). Commits `9f9db8c` / `b002fa2` / `7177515` /
`7add8eb`. This is the same feature that resolved #3005 (named props + free-form
values); witnessed by `fixtures/bugs/3005` + `TestCoverage_Bug3005` + golden.
forthcoming-features §17 marked ✅ DONE. 📖 Doc-site: document the
map-derived behaviour plus the new explicit markers/keywords.


<a id="a-2547"></a>

### #2547 — Unable to set a field value as empty string ""

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2547)

**✅ Fixed / lock (landed `64ceb83`, verified 2026-06-15).** The string
`example:`/`default:` quote-retention (quirk F8) is fixed: surrounding quotes are
stripped, so `example: ""` is the empty string and `example: "Foo"` is `Foo`.
`fixtures/bugs/2547` + `TestCoverage_Bug2547` and the flipped #2899 assertion are
green on `fix/backlog-lot1`.


<a id="a-2549"></a>

### #2549 — Example not working for imported Types

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2549)

**✅ Fixed / lock (verified 2026-06-15).** An example on a field whose type is an
imported named type (a `$ref`) is applied via `allOf: [{$ref}, {example}]` — no
longer dropped ("shows 0"). Locked by `fixtures/bugs/2549` (field of a
subpackage type) + `TestCoverage_Bug2549` + golden. The example value is carried
as the string `"210000"`; numeric coercion of example values is the separate
forthcoming-features §2.1 / #1268 concern.


<a id="a-2575"></a>

### #2575 — Custom Request Headers

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2575)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Custom request headers are
documented with a `swagger:parameters` field marked `in: header` — it emits a
`{in: header, type: string, ...}` parameter. Locked by `fixtures/bugs/2575` +
`TestCoverage_Bug2575` + golden. **Doc-site action:** a header-parameter example
in the parameters docs.


<a id="a-2588"></a>

### #2588 — Panic on parsing interface type definition

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2588)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** A struct embedding a named
type whose underlying type is an interface (`type B A; type C struct{ B }`) no
longer panics ("interface conversion: ast.Expr is *ast.Ident, not
*ast.InterfaceType"). A→object, B→`$ref` A, C→object. Locked by
`fixtures/bugs/2588` + `TestCoverage_Bug2588` + golden.


<a id="a-2592"></a>

### #2592 — Wrap generated spec with additional payload

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2592)

**✅ Fixed / lock + 📖 (re-triaged 2026-06-17).** In scope after all:
composing a response body from two types works via `// swagger:allOf` on the
embedded components of the BODY — a named `Combined{User,Token}` →
`allOf:[{$ref:User},{$ref:Token}]`, and an inline body with the same embeds →
that allOf directly on the response schema. (Embedding in the response *wrapper*
turns the field into a header instead — the embeds must live in the body.)
Locked by `fixtures/bugs/2592` + `TestCoverage_Bug2592` + golden. **Doc-site
action:** the allOf-composed response recipe (swagger:allOf embeds in the body).


<a id="a-2596"></a>

### #2596 — Overwrite Interface with specific type in response annotation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2596)

**✅ Doc-only / works-as-designed + 📖 (re-triaged 2026-06-17; doc landed on
`doc-site-update`).** The swaggo-style per-route override (`body:APIResponse{Data:
StatusReport}`) stays declined — codescan won't invent an undeclared type. But the
underlying need IS supported via a documented pattern: a **doc-only struct that
embeds the generic envelope and shadows the open `interface{}` payload with a
concrete type**. The shallower local field wins in both codescan and
`encoding/json`, so the spec gets a concrete `$ref` (`data`) while handlers keep
returning the generic envelope (the doc struct marshals identically). Documented
in the new "Documenting generic responses" shaping how-to, backed by the
golden-verified `docs/examples/shaping/genericenvelopes` package (commit
`dbef0bf`). **Doc-site action:** DONE — embed-and-shadow how-to + per-operation
specialisation; cross-links the `body:` sub-language. Related: #2419 (custom
swagger type for a struct field), #2592 (allOf-composed bodies).


<a id="a-2599"></a>

### #2599 — Custom type on models fields

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2599)

**✅ Fixed / lock + 📖 (verified 2026-06-14).** Works. A field-level
`// swagger:type string` on a custom Go type (`ID UUID`, an array under the
hood) renders the field as a bare `{type: string}` — no $ref, no strfmt needed.
Same mechanism answers #2404 and #2419. Locked by `fixtures/bugs/2599` +
`TestCoverage_Bug2599` + golden. 📖 Doc-site: document field-level `swagger:type`
override (the existing docs only show it on a named *type* declaration).


<a id="a-2618"></a>

### #2618 — How to add a description to the fields in the body part of JSON type in the swagger API？

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2618)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Field descriptions come from
the field's **godoc doc comment**, not from a custom `desc:"..."` struct tag
(codescan only reads the `json` tag for naming). A doc comment on each field —
including body/nested struct fields — already produces the property description.
**Doc-site action:** document that field descriptions are written as doc comments
above the field, and that arbitrary struct tags like `desc` are ignored.


<a id="a-2625"></a>

### #2625 — How to generate spec from code if have only one struct for all responses?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2625)

**✅ Works-as-designed + 📖 (verified 2026-06-14).** A `Data interface{}` field
renders as an empty (any-type) schema — correct: an untyped `interface{}` admits
any JSON. The reporter's real need (distinct payloads documented per endpoint
while reusing one envelope struct) is met by **response composition**
(`200: body:Envelope{Data: ConcreteType}`), not by one shared struct. No code.
📖 Doc-site: a "generic envelope" recipe — `interface{}` → any-schema, and the
`body:Wrapper{Field: Type}` composition syntax for per-endpoint payloads.


<a id="a-2626"></a>

### #2626 — Single line comment should never be parsed as title

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2626)

**✅ Fixed via opt-in option + 📖 (forthcoming §13 landed 2026-06-17, commit
`b67c2e0`).** The default first-sentence heuristic stays (a single-line doc
comment ending in a period → `title:`, otherwise → `description:`), matching
go-swagger convention. The OP's request is met by the new opt-in
`SingleLineCommentAsDescription` option, which treats single-line comments as
description-only across the schema/operations/routes/spec builders (parser
change in `internal/parsers/grammar/parser.go`). Covered by
`fixtures/enhancements/single-line-description` +
`TestCoverage_SingleLineCommentAsDescription` + `grammar/parser_test.go`;
documented in `grammar/README.md`. (Sibling #1118 — the title/description
inconsistency complaint — is the same §13 family and is now also satisfiable
with this knob.) 📖 Doc-site: document the title/description heuristic and the
new option.


<a id="a-2633"></a>

### #2633 — The `swagger generate spec` command does not run normally.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2633)

**✅ Fixed via fail-loud feature §8.2 (landed 2026-06-17, commit `d2479a4`).**
"Silently stops without `-w` on M1" — the underlying symptom was a degraded
package load (working-dir / package-discovery failure) producing an empty or
incomplete spec **silently**. §8.2 now runs `detectDegradedLoad` after
`packages.Load` and, on a degraded load, emits a located `scan.degraded-load`
diagnostic and aborts via `ErrDegradedLoad` instead of silently returning an
empty spec — so the failure is now visible and diagnosable. Witnessed by the
degraded-load case in `internal/scanner/scan_context_test.go`
(`ErrDegradedLoad` / `grammar.CodeDegradedLoad`). #2778 and #2874 are the same
degraded-load family (#2874 a likely dup). Was parked in the poison queue (now
emptied). 📖 Doc-site: `-w` / working-dir guidance and the new degraded-load
diagnostic.


<a id="a-2637"></a>

### #2637 — Cyclic type definition for defined types using the same name in spec generation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2637)

**✅ Fixed (landed 2026-06-17 — `feat/name-identity-cyclic-ref` merged into
`fix/backlog-lot1` at `9740c6d`; core commits `5bbd33a` name-identity
deconfliction + `d8f654c` cross-package leaf-type-name resolution).** A local
type defined from a same-named type in another package
(`type CreateDomainRequest mongo.CreateDomainRequest`) used to collide on the
short key and emit a definition whose body was a `$ref` TO ITSELF — invalid OAS
that hangs downstream codegen. The name-identity track now qualifies colliding
definition names by package leaf, so the local type and mongo's get distinct
names (`X2637CreateDomainRequest` / `MongoCreateDomainRequest`) and the local
definition's body is a `$ref` to the MONGO definition — a valid cross-type
reference, no self-`$ref`. Same family as #2783. Locked by `fixtures/bugs/2637`
+ `TestCoverage_Bug2637` (asserts no self-`$ref`, the cross-ref target, and a
`CodeCollidingModelName` diagnostic) + golden `bugs_2637_schema.json`. Design:
`.claude/plans/name-identity-cyclic-ref.md`.


<a id="a-2638"></a>

### #2638 — Improper handling of multiple variables on one line

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2638)

**✅ Fixed (`fix/go-swagger-2638`).** Root cause: `resolvers.ParseJSONTag`
defaulted the JSON name to `field.Names[0].Name`. A field group with several
names on one line (`R, G, B, A uint8`) expands in `go/types` to one
`*types.Var` per name, but all four share the same AST `*ast.Field`, so every
var resolved to the name `R` while keeping its own `goName` — last-write-wins
left a single property `R` carrying x-go-name `A`.

Fix: thread the per-field Go name (authoritative from `go/types`) into
`ParseJSONTag(afld, fld.Name())` and use it as the default name. A json rename
can only name a single field, so it is dropped for a multi-name group (each
member keeps its Go name) while `-`, `,omitempty` and `,string` still apply to
every member. The change is at the shared `resolvers` layer, so all three
field-walking builders (schema, parameters, responses) are fixed uniformly.

Fixture `fixtures/bugs/2638/api.go` covers the upstream `RGBA` shape plus a
`Mixed` model exercising a multi-name group with a json rename alongside
single-name fields. Unit coverage in `resolvers_test.go` (multi-name keeps each
goName; rename dropped but options apply; `-` ignores all). No golden drift on
the existing corpus. Documented in `schema/README.md` §method-mangler.


<a id="a-2639"></a>

### #2639 — Generate schemas only for referenced Models

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2639)

**✅ Works-as-designed / 📖 (verified 2026-06-13).** `-m` (`--scan-models`)
deliberately emits ALL `swagger:model` types; the default (without `-m`) already
emits only route-reachable models. The reporter (large shared library) wants
`-m` discovery **plus** pruning of unreferenced models — a middle ground, parked
as forthcoming-features.md §12. 📖 Doc-site: clarify `-m` semantics (all-models
vs reachable-only).


<a id="a-2651"></a>

### #2651 — Wrong binding when swagger:operation uses parameters and swagger:parameters bind to operations at same time

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2651)

**✅ Fixed / lock (verified 2026-06-16; merged `f9290da`).** An operation that
mixes inline parameters with swagger:parameters-bound ones now keeps them
distinct (correct binding). `fixtures/bugs/2651` + `TestCoverage_Bug2651` green
on `fix/backlog-lot1`.


<a id="a-2652"></a>

### #2652 — how to add complex example for a swagger:model

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2652)

**✅ Fixed (verified 2026-06-13).** An `example:` on a `$ref`'d field is
preserved on the allOf override arm (`allOf: [{$ref}, {example}]` — the #3125
shape); it used to be dropped. Locked by
`internal/integration/coverage_bug_2652_test.go` (fixture `fixtures/bugs/2652/`,
golden `bugs_2652_schema.json`). Residual (F8/coercion): a complex JSON example
on a $ref'd field is carried as a raw string, not a parsed object.


<a id="a-2655"></a>

### #2655 — Tags field ignored in Metadata when generating spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2655)

**✅ Fixed (`fix/go-swagger-2655`).** A `Tags:` block in swagger:meta is now a
recognized raw-block keyword (`KwTags`, meta-only) that parses into
`spec.Swagger.Tags` — name, description, nested `externalDocs` and per-tag
`x-*` vendor extensions all survive. Three layers changed:

- **grammar lexer** — `Tags:` joins the YAML-bodied raw blocks (preserves
  per-line indentation via the `Raw` view). New **indentation override**: in a
  YAML-bodied block, a same-family keyword indented *strictly deeper* than the
  head (e.g. `externalDocs:` under a `Tags:` list item, both meta-family) is
  absorbed as nested body, not treated as a sibling terminator. Flat raw blocks
  (`tos`/`consumes`/…) keep depth-agnostic termination — their indentation is
  cosmetic (petstore meta proves it).
- **yaml** — new `UnmarshalListBody` (sequence-shaped sibling of
  `UnmarshalBody`); shared dedent extracted to `normaliseBody`. `UnmarshalBody`
  left on `map[any]any` (decoding into `any` broke the operations map path via
  `YAMLToJSON`).
- **spec walker** — `KwTags` case unmarshals the list into `spec.Swagger.Tags`.

**Completed the tag story (follow-on, same branch):**

- **Route `Tags:` keyword** — `KwTags` is now legal in `CtxRoute`/`CtxOperation`
  too. On a route it is a plain string list (not objects), unioned+deduped onto
  `op.Tags` via `unionTags` (header-line tags + body `Tags:` names). Operations
  already worked via the wholesale YAML-body unmarshal — pinned by a test.
- **externalDocs coverage** — schema-level externalDocs was already wired (model
  + plain field via `handlers.schemaRawHandler`); the stale "unwired" note was
  wrong. The one real gap — a `$ref`'d field with a sibling `externalDocs:` — is
  now closed: `refOverrideCollector` collects it and lifts it onto the outer
  allOf compound (like description / `x-*`).

Tests: integration `TestCoverage_Bug2655` (meta tags + route/operation tags),
`TestCoverage_ExternalDocsObjects` (extended with field + `$ref`-field
externalDocs, golden regenerated), grammar `TestFixtures_Meta_TagsBlock`, yaml
`TestUnmarshalListBody_*`. READMEs (grammar/routes/schema) updated. Full suite +
Go lint + markdown lint clean.


<a id="a-2662"></a>

### #2662 — Same ref names while generating spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2662)

**♻️ Duplicate of #2783 (verified 2026-06-13).** Same-named structs in
different packages silently override each other, producing an invalid spec —
the same cross-package model-name collision as #2783 (seeded on
`fix/go-swagger-2783`, design-heavy: disambiguate names).


<a id="a-2663"></a>

### #2663 — How to document a body parameter in a POST request?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2663)

**✅ Doc-only / works-as-designed + 📖 (re-triaged 2026-06-17; the OP's repro was
buried in an attached image).** The correct idiom for a POST body: declare a
`swagger:parameters` struct holding **exactly one** field marked `// in: body`,
whose **Go type is the body schema**; give it a *named* type so the schema is
emitted as a reusable `$ref`. Witnessed by `fixtures/bugs/2663` +
`TestCoverage_Bug2663` (login/`CPF` domain from the issue): one `in: body`
parameter, `schema: $ref AppUserCredentials`, and the named-string `CPF` type
preserved as its own definition. The four proposals floated on the issue were
each tested:

- **OP's original** (each scalar field marked `in: body`) → emits **one body
  parameter per field**, i.e. *two* body parameters, which is **invalid Swagger
  2.0** (an operation may have at most one body parameter). This is the source
  of their confusion.
- **Inline anonymous `Body struct{…}`** → valid, single body param, but an
  **inline** schema (no reusable `$ref`); the param name also leaks the field
  name when there is no `json` tag.
- **`type _ struct{ Body NamedType }`** → works identically to the recommended
  form (clean `$ref`); the blank `_` wrapper name is irrelevant to output, but a
  real (unexported) name is clearer/greppable.
- **Unexported `Body struct{ username … }`** → unexported fields emit **no
  properties**; as written (no backtick tags) it does not even compile.

**Doc-site action:** document the one-field/named-type body-parameter idiom and
the "at most one `in: body`" rule; show the named-type → `$ref` form as the
recommendation.


<a id="a-2687"></a>

### #2687 — Ignore kubebuilder annotations

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2687)

**✅ Fixed (2026-06-14).** Tool directive markers (`+kubebuilder:…`,
`+genclient`, `+k8s:…`) used to leak into model & property descriptions. The
grammar lexer now recognises `+marker`-style lines (`+` + a letter) via
`isDirectiveMarker` and drops them from the prose surface, alongside the
existing Go `//go:`/`//nolint:` directive handling.

Key design point: the drop runs at **Stage 3** (`classifyProse`), not at line
classification — the inline `swagger:route` parameters grammar uses `+name:` as
a parameter separator (#3100), which matches the marker shape, and by Stage 3
that route body has already been folded into its keyword token by
`accumulateBodies`, so only loose prose lines are filtered. Requiring a letter
after the `+` keeps ordinary prose (`+1 …`, markdown `+` bullets) intact.

This also closes **#3007's residual** (the `+kubebuilder:default:=false` marker
leaking into the field description): `coverage_bug_3007_test.go` flipped from
asserting the marker present to asserting the clean description, golden
`bugs_3007_schema.json` regenerated. Verified by `coverage_bug_2687_test.go`,
`coverage_bug_3007_test.go`, `coverage_bug_3100_test.go` (separator preserved),
plus grammar unit tests `TestLexer_DirectiveMarkerPredicate` /
`TestLexer_DirectiveMarkersDroppedFromProse`. Branch `fix/go-swagger-2687`.


<a id="a-2701"></a>

### #2701 — In path parameter for an embedded struct is ignored and thus default to  in query parameter

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2701)

**✅ Fixed (2026-06-14).** `in: path` (and `required: true`) on an embedded
struct field in swagger:parameters was ignored — the promoted fields defaulted
to `in: query`, invalid for the route's path templates. The parameters builder
recursed into the embed by *type* only, so the embed's own doc-comment
annotation was never read.

Fix (generalized with the maintainer — this is a cross-builder rule, not a
parameters quirk): a shared kernel `common.EmbedInheritance` +
`Builder.ReadEmbedInheritance` reads the embed's `in:`/`required:` directives;
each field-walking builder threads it through its own embed recursion with
save/restore (siblings unaffected, nesting accumulates) and applies it as a
fallback — a promoted member's own directive always wins. Per-builder:

- **parameters** apply `in:` + `required:` (the #2701 case);
- **schema** applies `required:` (adds promoted props to the enclosing object's
  required list); it has no `in:` concept;
- **responses** apply `in:` (body/header routing); response **bodies** inherit
  `required:` through the schema builder (a body is built there); OAS2 response
  headers carry no `required`.

The `in:` line-scan was de-duplicated into `common.ScanInLocation` (was copied
in parameters + responses). Exportedness stays per-field (the embed recursion is
NOT gated on the embed's own exportedness): exported fields reached through an
*unexported* embed still promote (Go promotes them; reachable on the outer
type), while unexported fields never surface at any depth.

Verified by: `coverage_bug_2701_test.go` (parameters — vendorID/urcapID/version
+ `trace` via an unexported `auditInfo` embed, all required `in: path`;
unexported `secret`/`token` absent); `coverage_embed_inheritance_test.go`
(schema model + response body `required:` propagation; `in: body` embed routing;
unexported member excluded); and a `common` unit test for `ScanInLocation`. Zero
golden churn on existing fixtures. Documented in the `common`, `schema`,
`parameters`, and `responses` READMEs. Branch `fix/go-swagger-2701` (also
carries the empty triage commits for duplicates #2762 N/A and #2791
works-as-designed).


<a id="a-2746"></a>

### #2746 — Strfmt: array of UUID

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2746)

**✅ Fixed (verified 2026-06-13).** `[]UUID` (swagger:strfmt uuid) →
`items: {type: string, format: uuid}` — strfmt now applies to array items, not
just scalars. Locked by `internal/integration/coverage_bug_2746_test.go`
(fixture `fixtures/bugs/2746/`, golden `bugs_2746_schema.json`).


<a id="a-2761"></a>

### #2761 — generate: allOf in response doesn't use $ref

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2761)

**✅ Works-as-designed / 📖 (verified 2026-06-13).** A response-body
`swagger:allOf` emits `allOf: [{$ref}, …]` when the embedded base is a
swagger:model (a definition to reference). The reporter embedded a
swagger:response (no definition) → its fields were inlined. Locked by
`internal/integration/coverage_bug_2761_test.go` (fixture `fixtures/bugs/2761/`,
golden `bugs_2761_schema.json`). 📖: allOf $ref needs a swagger:model base.


<a id="a-2762"></a>

### #2762 — codescan tests fail on go 1.18

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2762)

**✅ N/A (verified 2026-06-13).** v1's codescan unit tests assumed a fixed map
field-order, which changed in go 1.18. Not applicable here: our integration
tests compare against deterministically-marshaled golden JSON (stable key
order), so there is no map-order assumption to break. Recorded via an empty
commit.


<a id="a-2791"></a>

### #2791 — swagger:parameters not working as defined in docs

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2791)

**✅ Works-as-designed / 📖 (verified 2026-06-13).** The deeply nested
`[][][]string` query param generates correctly (in:query, nested items,
min/max/unique). The docs example is invalid OAS2 — `example:` is forbidden on
a non-body param, and `collection format: pipe` should be `pipes`; codescan
emits the annotations as written. 📖 Doc-site: fix the docs example. (Hardening
— warn/drop example on simple params, validate collectionFormat — noted, not
pursued.)


<a id="a-2799"></a>

### #2799 — swagger generator should keep the format of the API description.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2799)

**✅ Fixed (verified 2026-06-13).** A markdown bullet list in the description is
preserved verbatim (`- keep markdown list\n- in the description`) — the
grammar2 §4.2 bullet-dash preservation. Locked by
`internal/integration/coverage_bug_2799_test.go` (fixture `fixtures/bugs/2799/`,
golden `bugs_2799_schema.json`). Distinct from #3211's markdown-table
leading-pipe stripping (still open).


<a id="a-2801"></a>

### #2801 — Spec is not generated if generic struct declaration is not in the same file as swagger:parameters

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2801)

**✅ Fixed (verified 2026-06-13).** A generic struct declared in a different
file from its swagger:parameters instantiation now resolves (Request emitted);
go/packages loads the whole package. Locked by
`internal/integration/coverage_bug_2801_test.go` (fixture `fixtures/bugs/2801/`,
golden `bugs_2801_schema.json`).


<a id="a-2802"></a>

### #2802 — Error generating spec if swagger:model is a generic struct

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2802)

**✅ Fixed (verified 2026-06-13).** An instantiated generic struct
(`WrappedRequest[Request]`) no longer panics ("can't determine refined type
T"); the type argument resolves (`Body.Body → $ref: Request`) and a free type
param is skipped with a warning. Locked by
`internal/integration/coverage_bug_2802_test.go` (fixture `fixtures/bugs/2802/`,
golden `bugs_2802_schema.json`).


<a id="a-2804"></a>

### #2804 — Should swagger:parameters support map[string][]string?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2804)

**✅ Fixed (2026-06-13).** A `map[string][]string` swagger:parameters field
used to PANIC (nil-ptr in `parameterBuilder.buildFromField` → `Schema.Typed`) —
a map is not representable as an OAS2 simple (query/formData/path/header) param.

Fix on `fix/go-swagger-2804`: the guard covers **both** OAS2 SimpleSchema
targets, applying the identical `typable.In() != inBody` rule in
`buildFromFieldMap`:

- **Parameters** (`internal/builders/parameters`): a map in a non-body location
  (query/formData/path/header) used to deref a nil SimpleSchema target and
  panic. `buildFromFieldMap` now returns the internal `errUnrepresentableParam`
  sentinel; `processParamField` records a located
  `validate.unsupported-in-simple-schema` warning and skips the field.
- **Response headers** (`internal/builders/responses`): the same map field did
  NOT panic (responseTypable.Schema() always returns the *body* schema) but
  silently wrote `object` onto the response body and left the header untyped.
  `buildFromFieldMap` now returns `errUnrepresentableHeader`;
  `processResponseField` records the same warning and skips the header.

Body maps (object + additionalProperties) are unaffected in both. Verified by
`internal/integration/coverage_bug_2804_test.go` (no panic; param + header both
skipped; response body schema not corrupted; both diagnostics asserted). The
fixture (`fixtures/bugs/2804/api.go`) carries both a map query param and a map
response header. Also a live witness for the §8.1 panic-recovery feature
(forthcoming-features.md) — the recover boundary, once landed, will turn any
remaining such panic into the same located diagnostic.


<a id="a-2778"></a>

### #2778 — Generating Empty Spec File

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2778)

**✅ Fixed via fail-loud feature §8.2 (landed 2026-06-17, commit `d2479a4`).**
"Empty spec unless `sudo`" — a permission/environment-degraded package load that
used to yield an empty spec **silently**. Same degraded-load family as #2633 /
#2874: §8.2's `detectDegradedLoad` now emits a located `scan.degraded-load`
diagnostic and aborts via `ErrDegradedLoad` rather than silently producing an
empty spec, making the permission failure visible. Witnessed by the
degraded-load case in `internal/scanner/scan_context_test.go`. Was parked in the
poison queue (now emptied). 📖 Doc-site: empty-spec troubleshooting (permissions,
working dir) and the new degraded-load diagnostic.


<a id="a-2783"></a>

### #2783 — Models get mixed when using structs from several packages

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2783)

**✅ Fixed (landed 2026-06-17 — `feat/name-identity-cyclic-ref` merged into
`fix/backlog-lot1` at `9740c6d`; core commits `5bbd33a` name-identity
deconfliction + `d8f654c` cross-package leaf-type-name resolution).** Two
packages each declaring `swagger:model Test` used to collide on the short key
"Test" and SILENTLY MERGE — the union of both structs' fields, last-package-wins,
non-deterministic (which `x-go-package`/`required` won was undefined). The
name-identity track now gives each its own identity: definitions are deep-keyed
by package while the leaf collides, so `b.Test` → `BTest` (only b's field) and
`c.Test` → `CTest` (only c's field) emit as two distinct definitions alongside
`TestResponseBody` — three definitions, deterministic, no merged bare `Test`
key. Locked by `fixtures/bugs/2783` + `TestCoverage_Bug2783` (asserts the two
Tests stay distinct and neither absorbs the other's field) + golden
`bugs_2783_schema.json`. Design: `.claude/plans/name-identity-cyclic-ref.md`.


<a id="a-2837"></a>

### #2837 — Responses defined in routes break with go 1.19 formatting in 1.30.*

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2837)

**✅ Fixed (verified 2026-06-13).** A route's `+ name:` parameter bullet and its
gofmt-rewritten `- name:` form both parse to the same parameter — the 1.30
gofmt breakage is gone. Locked by
`internal/integration/coverage_bug_2837_test.go` (fixtures
`fixtures/bugs/2837{a,b}/`, golden `bugs_2837_schema.json`).


<a id="a-2838"></a>

### #2838 — can not generate swagger spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2838)

**✅ Fixed — Go toolchain mismatch + 📖 (re-triaged 2026-06-17; empty
`* contributes` commit, no witness).** The empty 3-line spec (`info:{} paths:{}
swagger:"2.0"`) despite swagger:meta comments was a **Go version / toolchain
issue** — the installed Go couldn't properly load the project's packages —
resolved by adding the `toolchain` directive to `go.mod` (pinning a compatible
toolchain) (per Fred's recollection of the case). Not a scanner-logic bug; the
basic swagger:meta path works (meta/v1–v4 fixtures). No CI witness (toolchain /
environmental). 📖 Doc-site: note that a Go/toolchain mismatch can yield an empty
spec and that the `toolchain` directive resolves it.


<a id="a-2846"></a>

### #2846 — Wrong format yaml format for enums outside body

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2846)

**♻️ Duplicate of #2922 / ✅ resolved (verified 2026-06-14).** The malformed
`description: |4-` YAML for an enum on a non-body parameter comes from the same
folded multi-line const mapping as #2922; it is resolved by the same
`SkipEnumDescriptions` knob (mapping kept out of the description). Fixed in the
#2922 commit (which carries `* fixes #2846`); no separate fixture.

<a id="a-2860"></a>

### #2860 — Models do not show fields from the struct.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2860)

**✅ Fixed (verified 2026-06-13).** The petstore models emit their full fields,
required sets, and nested item schemas under `-m` (the v0.30 fields-less output
is gone). Already covered by the petstore goldens (`petstore_schema_Order.json`
etc.); contribution recorded via an empty commit.


<a id="a-2871"></a>

### #2871 — Question: dynamic examples

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2871)

**✅ Fixed / lock + 📖 (re-triaged 2026-06-17; struct path landed v0.36
2026-06-23).** The OP's literal "dynamic examples" (per-request contextual
values) stays out of scope, but the reasonable solution IS supported: distinct
**per-operation, per-response examples** (by mime type). Multiple operations
reuse one example-bearing model, each response code carrying its own example,
all different from the model's. Locked by `fixtures/bugs/2871` +
`TestCoverage_Bug2871` + golden (the `swagger:operation` YAML path).

The **struct-based `swagger:response` path** now also supports response-level
`examples:` (a mime-keyed YAML map → `Response.examples`), landed v0.36 on
`feat/feature-v0.36` — new `examples` grammar keyword (CtxResponse) parsed in
`responses/walker.go`; locked by `fixtures/enhancements/response-examples-by-mime`
+ `TestCoverage_ResponseExamplesByMime` + golden. See `features/response-examples-by-mime.md`.
**Doc-site action:** per-response `examples:` (by mime) on BOTH the
swagger:operation YAML body and the struct `swagger:response`.



<a id="a-2872"></a>

### #2872 — "ExternalDocs" are not generating the 2.0 spec on swagger:meta

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2872)

**✅ Fixed (2026-06-13) on `fix/go-swagger-2872`.** `ExternalDocs` in
swagger:meta was NOT emitted (`spec.ExternalDocs` stayed nil) — `KwExternalDocs`
existed in the grammar (`asRawBlock`) but was wired into no builder. Fix: added
a `KwExternalDocs` case to `spec/walker.go dispatchMetaYAMLBlock` (alongside
`securityDefinitions`) that YAML-unmarshals the body into
`spec.ExternalDocumentation` and sets the top-level `swspec.ExternalDocs`; an
empty/blank block is skipped (no useless `externalDocs: {}`). Tests: bugs/2872
flipped to assert description+url; enhancements/external-docs covers the
`external docs` alias + url-only and the empty-block skip. 📖 Doc done:
docs/keywords.md §externalDocs now states emission is meta-only (route/
operation/schema grammar-legal but not yet emitted). Companion feature:
externalDocs on non-meta OAIv2 objects (schema/operation/tag) —
forthcoming-features.md §9.


<a id="a-2874"></a>

### #2874 — go-swagger silently stops parsing `swagger:allOf` when `GOROOT` is not set

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2874)

**✅♻️ Fixed via fail-loud feature §8.2 / likely dup of #2633 (landed 2026-06-17,
commit `d2479a4`).** `swagger:allOf` itself works (see #2875); the reported
problem was a GOROOT-unset **silent-stop** — a degraded package load producing an
incomplete spec with no warning. §8.2 is this issue's named origin: `detectDegradedLoad`
after `packages.Load` now emits a located `scan.degraded-load` diagnostic and
aborts via `ErrDegradedLoad`, so a GOROOT/env-degraded scan fails loudly instead
of silently dropping definitions. Same docker/env family as #2633 (likely dup) and
#2778. Witnessed by the degraded-load case in
`internal/scanner/scan_context_test.go`. forthcoming-features §8.2 marked ✅ DONE.
Was parked in the poison queue (now emptied).


<a id="a-2875"></a>

### #2875 — AllOf member does not generate an external $ref object

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2875)

**✅ Fixed (verified 2026-06-13).** A `swagger:allOf` embedded member now emits
`allOf: [{$ref: …}]` to the embedded model's definition (was inline). Locked by
`internal/integration/coverage_bug_2875_test.go` (fixture `fixtures/bugs/2875/`,
golden `bugs_2875_schema.json`).


<a id="a-2886"></a>

### #2886 — Invalid memory address or nil pointer (SIGSEGV)  => swagger generate spec

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2886)

**✅ Fixed via hardening feature §8.1 (landed 2026-06-17, commit `bc568c7`).**
The exact SIGSEGV (`responses.go` `Schema.Typed` on nil) isn't reproducible
without the reporter's source, and the restructured responses builder already
guards that specific path — but the general class is now handled: forthcoming
§8.1 wraps the six spec build loops (and a coarse `Run` backstop) in a per-decl
`guard` that `recover()`s any builder panic, emits a *located*
`scan.internal-panic` diagnostic (file:line + label of the offending
declaration) and aborts via `ErrInternalPanic` — no more raw stack trace, no
silent corruption. Witnessed by `TestBuilder_guard_RecoversPanicWithLocatedDiagnostic`
(`internal/builders/spec/panic_guard_test.go`). forthcoming-features §8.1 marked
✅ DONE.


<a id="a-2897"></a>

### #2897 — with go1.20 swagger missing definitions and refs

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2897)

**✅ Works-as-designed (verified 2026-06-13).** The reporter's example (a
swagger:response with a struct Body) produces the definition + `$ref` correctly
in the current scanner. The failure was environmental (a debian-repo binary vs
a self-built one / go-version mismatch), not a scanner defect. No fixture
(covered by #1109 / #2907). Contribution recorded via an empty commit.


<a id="a-2898"></a>

### #2898 — Output of []any is changed in 0.30 so that it is interpreted as slice of strings

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2898)

**✅ Works-as-designed / 📖 Need doc (verified 2026-06-13).** `[]interface{}`
emits `items: {}` (an empty schema = "any type"), which is correct —
`{type: object}` (pre-0.30) wrongly restricted elements to objects. The change
is a correctness improvement, not a regression. 📖 Doc-site: explain the
rationale.


<a id="a-2899"></a>

### #2899 — Example not being added to schema for body string params

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2899)

**✅ Fixed (verified 2026-06-13).** An `example:` on an inline primitive body
parameter is now carried onto the parameter's schema. Locked by
`internal/integration/coverage_bug_2899_test.go` (fixture `fixtures/bugs/2899/`,
golden `bugs_2899_schema.json`). Known residual: quoted string examples retain
their surrounding quotes — tracked as quirk **F8** (doc-site-quirks.md).


<a id="a-2907"></a>

### #2907 — Go-Swagger not generating properties in yaml file

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2907)

**✅ Fixed (verified 2026-06-13).** A response body that is a slice of a
cross-package model (`Body []data.Movie`) now resolves to `items: {$ref:
#/definitions/Movie}` with the model emitted (with/without `-m`); v0.30.5's
`items: {}` is gone. Locked by `internal/integration/coverage_bug_2907_test.go`
(fixture `fixtures/bugs/2907/{data,api}`, golden `bugs_2907_schema.json`).


<a id="a-2909"></a>

### #2909 — Regular cannot generate swagger automatically

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2909)

**✅ Fixed (2026-06-13) on `fix/go-swagger-2909`.** A route path with an
inline regex segment (gorilla/chi style, `{id:[0-9]+}`) was SILENTLY dropped —
`rxPath`'s alphabet lacks `[`/`]`, so the swagger:route line failed to match and
no path was emitted, with no diagnostic. Fix: `parsers.stripPathParamRegex`
(brace-depth-aware, handles `{id:[0-9]{2,4}}`) rewrites `{name:regex}` →
`{name}` on the line BEFORE the rxRoute/rxOperation match, so the route now
emits as `/items/{id}` (no `rxPath` alphabet change needed). New
`ParsedPathContent.{Pos,StrippedParams}` carry the strip info; the routes AND
operations builders call shared `common.Builder.WarnStrippedPathRegex`, which
warns (`parse.invalid-annotation`) that OpenAPI 2.0 path templating follows
RFC 6570 URI Template **Level-1 expansion only** (per fred: go-openapi's
runtime path-templating middleware supports Level-1 only, so the constraint is
genuinely unsupported, not silently kept). Tests: bugs/2909 flipped to assert
the fix; enhancements/path-regex-stripping covers route + operation + nested-
brace + plain control; unit table-test for the stripper. 📖 Doc done:
docs/annotations.md `swagger:route` path section now documents RFC 6570 Level-1
+ the strip behaviour.


<a id="a-2912"></a>

### #2912 — Problem of swagger generate ,about struct tag  "json/form"

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2912)

**✅ Works-as-designed / 📖 Need doc (verified 2026-06-13).** Parameter names
derive from the `json:` tag (then the Go field name); the `form:` tag is not
consulted. A field with only `form:"sort_key"` is named `SortKey`; adding
`json:"sort_key"` names it `sort_key`. No code change (no fixture — json-tag
naming is exercised by the named-struct-tags corpus). 📖 Doc-site: parameter
names come from json tags; the form-tag workaround.


<a id="a-2917"></a>

### #2917 — classifier: unknown swagger annotation "extendee" when importing github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2917)

**✅ Fixed (verified 2026-06-13).** A grpc-gateway generated comment containing
`…openapiv2_swagger:extendee…` no longer triggers `classifier: unknown swagger
annotation` — the current regex matches `swagger:` only at start-of-line or
after whitespace/slash. Locked by
`internal/integration/coverage_bug_2917_test.go` (fixture `fixtures/bugs/2917/`,
golden `bugs_2917_schema.json`).


<a id="a-2922"></a>

### #2922 — [spec generation] enum description: superfluous name&values

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2922)

**✅ Fixed / 📖 Need doc (2026-06-13).** The enum const-name mapping was folded
into the `description` AND duplicated in `x-go-enum-desc`, at every target that
folds it in — schema properties (in: body) and non-body parameters (in: query).

Fix on `fix/go-swagger-2922`: new Options knob **`SkipEnumDescriptions`**
(default `false` = current append, backward compat; `true` = description left as
authored prose, mapping rides `x-go-enum-desc` only). The three append sites
(schema model-decl block, schema struct-field, parameter field) now route
through one shared gate, `resolvers.AppendEnumDesc(desc, ext, skip)`, so the
knob behaves identically everywhere. Independent of `SkipExtensions` (both set ⇒
mapping suppressed everywhere).

**Response headers — the third SimpleSchema target (completed):**
headers never fold the mapping into the description, so the knob doesn't touch
them. They now DO expose `x-go-enum-desc`: the earlier limitation was that
go-openapi/spec's `Header.MarshalJSON` serialized only CommonValidations +
SimpleSchema + HeaderProps and dropped the embedded `VendorExtensible` (a
10-year-old marshal/unmarshal asymmetry — `UnmarshalJSON` read extensions since
2016 but `MarshalJSON` never emitted them). Fixed upstream in
**go-openapi/spec v0.22.6** (go-openapi/spec#277, mirroring the `Items`
marshaler); `responseTypable.WithEnumDescription` is now wired to set
`x-go-enum-desc` on the header (gated on `SkipExtensions`, mirroring
`schema.Typable`). The extension is present in both knob modes (headers never
folded into the description, so `SkipEnumDescriptions` doesn't change them).
While first probing this we also found and fixed a real, output-affecting bug:
`responseTypable.WithEnum` passed the slice instead of spreading it (`values`
vs `values...`), emitting malformed double-nested `enum: [[FIRST, SECOND]]` on
every enum-typed response header; siblings (param/schema/items) were already
correct.

Verified by `internal/integration/coverage_bug_2922_test.go` (two sub-tests:
default + skip; schema property, query parameter, AND response header asserted —
the header now asserts `x-go-enum-desc` present in both modes) with goldens
`bugs_2922_schema.json` and `bugs_2922_skip_enum_desc_schema.json` (the goldens
ARE the marshaled JSON, so they prove the header extension survives marshal
end-to-end). The fixture adds an enum-typed `swagger:response` header proving
the flat-enum fix and the extension. Requires spec >= v0.22.6 (carried on
`fix/backlog-lot1`). 📖 Doc-site: the new knob.


<a id="a-2924"></a>

### #2924 — is it possible to specify "x-go-type" extension when swagger spec generates?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2924)

**✅ Fixed via opt-in option + 📖 (forthcoming §20 landed 2026-06-17, commit
`de306ca`).** Emitting an `x-go-type` vendor extension from code→spec is now
available through the new opt-in `EmitXGoType` option, which stamps `x-go-type`
(originating package + type) on emitted definitions — reusing the same emit seam
as the existing `x-go-name` / `x-go-package` extensions. Off by default (so no
golden drift); aids round-tripping when on. Covered by
`fixtures/enhancements/emit-x-go-type` + `TestCoverage_EmitXGoType`; documented
in `schema/README.md`. 📖 Doc-site: document the `x-go-*` extensions codescan
emits and the new option.


<a id="a-2942"></a>

### #2942 — how can i generate spec with a simple  string response

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2942)

**✅ Fixed (2026-06-13) on `fix/go-swagger-2942`.** No easy path to a
primitive string response: `200: string` was treated as a response *name*
(dropped, the #1109 resolution order) and a swagger:response with a primitive
`Body string` emitted no schema. Two root causes, both fixed:

- **swagger:response primitive body** — `responseTypable.Typed` always wrote
  to the header; a body field's primitive type landed on a header that body
  responses discard. Now branches on `in == inBody` → writes to the body
  schema (mirrors `SetRef`). `Body string` → `schema: {type: string}`.
- **`200: body:string` route response** — the route response builder only
  resolved refs. New `primitiveBodySchema` helper emits a typed schema for the
  scalar primitives `responseBodyPrimitives` = {string, number, integer,
  boolean}, array-wrapped per `[]` (`body:[]string`), wired into the `body:`
  tag path. The reserved keywords `responseBodyReservedTypes` = {array, object,
  file, null} draw a diagnostic (use `[]T` / a model name; `file` dropped per
  fred — extra produces/context checks, low value); anything else after `body:`
  is a model `$ref`. **Decision (fred):** the bare/untagged `200: string` form
  is intentionally NOT supported — an untagged token is a response/model NAME,
  and reading it as a type makes the syntax ambiguous; `body:<type>` is the
  unambiguous supported path (no Go type name carries a `:`). A bare primitive
  spelling drops with a diagnostic pointing at the `body:` form.

Tests: bugs/2942 flipped to assert both fixes (route uses `200: body:string`);
enhancements/primitive-response covers tagged primitive / array-of-primitive /
swagger:response array body / model-ref control / dropped bare-primitive + hint.
Zero golden churn. 📖 Doc done: docs/annotations.md
(swagger:response body may be primitive) + docs/sub-languages.md (Responses
`body:` tag accepts primitives; untagged token stays a name).


<a id="a-2959"></a>

### #2959 — Unable to provide security definitions

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2959)

**✅ Fixed (verified 2026-06-13).** Tab-indented `SecurityDefinitions` in
swagger:meta — the v0.30.5 crash trigger ("found character that cannot start
any token") — now parses cleanly. Already covered by the existing meta
fixtures (`fixtures/goparsing/meta/v3` uses tab-indented SecurityDefinitions;
`enhancements_meta_securitydefs_duplicate_keys.json`), so no new fixture —
contribution recorded via an empty commit. **Residual:** the
gofmt-*canonical* form (a column-0 key + the blank `//` line gofmt inserts +
tab-indented children) still errors — tracked as quirk **F7** in
[`doc-site-quirks.md`](doc-site-quirks.md), branch `fix/scanner-gofmt-meta-yaml`.
So #2959 is resolved for non-gofmt'd meta forms; the gofmt residual lands with F7.


<a id="a-2961"></a>

### #2961 — Improper parsing of uint enums

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2961)

**✅ Fixed (verified 2026-06-13).** An `enum:` on a `uint`/`uint64` field is
coerced against the integer schema type and emitted as numbers (`[1,2,3]`),
not strings. Locked by `internal/integration/coverage_bug_2961_test.go`
(fixture `fixtures/bugs/2961/`, golden `bugs_2961_schema.json`).


<a id="a-2963"></a>

### #2963 — Unable to reference models in referenced module

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2963)

**✅ Fixed — observed only + 📖 (re-triaged 2026-06-17; local repro, no CI
witness).** Cross-module model resolution via a `replace` directive works. A
local multi-module repro — an app module with `replace example.com/dep =>
../depmod` and a body parameter typed as `dep.Widget` — produced a complete
`Widget` definition (`id`/`name`, `x-go-package: example.com/dep`) and a clean
`$ref`; the OP's "fields show null" does **not** reproduce. Modern
`golang.org/x/tools/go/packages` resolves replace-directed modules. Observed on a
throwaway local worktree; deliberately **not** locked as a CI witness (a
multi-module replace harness is out of scope for the suite — observed only). 📖
Doc-site: cross-module / replace-directive scanning works.


<a id="a-2980"></a>

### #2980 — Should I expect embedded structs to be supported in generated spec files?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2980)

**✅ Works-as-designed + 📖 (verified 2026-06-15).** Embedded (anonymous) structs
ARE supported: their fields are promoted into the composed model (cross-links
#2027, #413, and #2038 for the json-tagged-embed nesting case). **Doc-site
action:** document embedded-struct composition (promotion; allOf where
applicable; tagged-embed nesting tracked as #2038).


<a id="a-2985"></a>

### #2985 — Need minAttributes and maxAttributes in the swagger:model annotation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2985)

**✅ Fixed (2026-06-13) on `fix/go-swagger-2985`.** `minProperties` /
`maxProperties` (the reporter's "minAttributes/maxAttributes") were NOT
grammar keywords, so the annotation lines were swallowed into the model
description and no validation was emitted on the object schema. Fix: added
both as `CtxSchema`-only `asInt()` grammar keywords (object validations have
no SimpleSchema equivalent), wired through `SchemaValidations.SetMin/MaxProperties`
in both schema-dispatch paths (`schemaIntegerHandler` + `walker.go onInteger`),
mirroring `minItems`/`maxItems`. The `validations.keywordTypeRules` object
shape-rule was already present. Golden + test flipped to assert
`minProperties:1`/`maxProperties:10` and no description leak; keyword docs
updated.

Follow-on work landed on the same branch:

- **`patternProperties`** added as a sibling object keyword (`asString()`,
  `CtxSchema`-only, object shape-rule). String argument is a regex; each
  line adds a `regex → {}` entry to `schema.patternProperties` and the
  regex gets the same RE2-hygiene check as `pattern` (`CodeInvalidAnnotation`
  on a non-compiling value, but the value is preserved — never dropped
  silently). Fixture `enhancements/pattern-properties` + golden + test.
- **Wrong-context diagnostics wired** (these were silently mis-handled
  before): (a) **non-object model** — top-level scalar models dispatch the
  doc block before the Go type is resolved, so the inline `checkShape` saw
  an empty type and applied object/array/string/number keywords anyway. New
  `handlers.RecheckSchemaShape`, called from `schema.Build` after
  `buildFromDecl`, strips validations illegal for the resolved type and
  emits `CodeShapeMismatch`. Chosen as a post-hoc strip over a dispatch
  reorder to avoid changing `default:`/`enum:` coercion timing (zero golden
  churn). (b) **SimpleSchema sites** — the param/header/items dispatchers
  silently dropped full-Schema-only object keywords; `handlers.Integer` now
  takes a `diag` and, with `UnsupportedSimpleSchemaString`, emits
  `CodeUnsupportedInSimpleSchema` for keywords whose grammar contexts are
  `CtxSchema`-only (predicate `isFullSchemaOnly` — correctly excludes
  param-legal keywords like `in:`). Fixture `enhancements/object-keywords-context`
  + golden + test cover all three contexts (kept / shape-mismatch / simple-schema).


<a id="a-3005"></a>

### #3005 — additionalProperties are lost when generating spec from code

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3005)

**✅ Fixed / lock (landed 2026-06-17).** A map field tagged `json:"-"`, intended
to carry the model's `additionalProperties`, was dropped — only the named
properties were emitted. The fix is the new explicit `swagger:additionalProperties`
type-level marker (`9f9db8c`): the muted `json:"-"` map stays muted, and the
marker supplies the `additionalProperties` value schema alongside the named
`field1`. Shipped with the wider additionalProperties surface — the
`additionalProperties:` field keyword (`b002fa2`), the typed `swagger:patternProperties`
marker (`7177515`), and the SimpleSchema safeguard (`7add8eb`). Locked by
`fixtures/bugs/3005` + `TestCoverage_Bug3005` + golden `bugs_3005_schema.json`.


<a id="a-3007"></a>

### #3007 — [Bug]generate spec error

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3007)

**✅ Fixed (verified 2026-06-13; residual closed 2026-06-14).** A non-swagger
kubebuilder marker (`+kubebuilder:default:=false`) no longer aborts the scan
(`strconv.ParseBool parsing "=false"`). The earlier residual — the marker being
absorbed into the field description rather than ignored — is now closed together
with #2687: the grammar lexer drops `+marker`-style lines from prose (see
[a-2687](#a-2687)). Locked by `internal/integration/coverage_bug_3007_test.go`
(fixture `fixtures/bugs/3007/`, golden `bugs_3007_schema.json`, now asserting the
clean `"Enabled flag."` description).


<a id="a-3013"></a>

### #3013 — How to set a example value for array/string response type?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3013)

**✅ Fixed (2026-06-13) on `fix/go-swagger-3013`.** A `swagger:response`'s
`example:` was dropped on non-struct bodies. Two distinct root causes, both
fixed:

- **Case A — top-level array/scalar response type** (the seeded fixture, e.g.
  `type getNtpServersResponse []string` with `example: [...]`):
  `responses.applyBlockToDecl` only took the decl prose, ignoring all keywords.
  Fix: after `buildFromType`, dispatch the response decl-comment's schema
  keywords (example/default/validations) onto the non-struct body schema via
  `handlers.DispatchSchemaLevel0` (guarded: skip struct bodies and `$ref`
  bodies). `CoerceValue` already turns the JSON-array value into `[]any`.
- **Case B — struct `Body` field example** (surfaced while probing; a
  #2942-family routing bug, not in the seeded fixture): a `Body` field's
  `example:`/validations went through `applyBlockToHeader` to the header, which
  body responses discard — so even a scalar `Body string` example was dropped
  (the backlog's "scalar field examples work (#3035)" refers to non-body
  fields). Fix: `processResponseField` now dispatches a body field's schema
  keywords onto the body schema (`resp.Schema`) instead of the header.

Tests: bugs/3013 flipped (top-level array example). enhancements/
response-toplevel-example (array own-paragraph + scalar trailing) and
response-bodyfield-example (struct Body array + scalar) lock both cases. Zero
golden churn. No 📖 (bug fix to the existing `example:` keyword).


<a id="a-3035"></a>

### #3035 — Example spec for swagger:response does not produce example output

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3035)

**✅ Already fixed (verified 2026-06-12).** A `swagger:response` with an
anonymous inline `Body struct{}` now emits a full object schema with its
`required` set, property descriptions, and per-field `example` (the subject
of the issue — examples in responses work). Confirmed both with and without
a blank line after the body field's leading prose.

**Documented delta vs the reporter's expected snippet 3:** the schema-level
`description: The error message` (sourced from the `Body` field's leading
comment) is NOT emitted. No codescan path surfaces a body field's prose as
the schema description (verified against all response goldens), and the
blank line does not enable it — so this is consistent behaviour, not a
regression. Surfacing it would be a separate enhancement. Locked by
`internal/integration/coverage_bug_3035_test.go` (fixture
`fixtures/bugs/3035/`, golden `bugs_3035_schema.json`).


<a id="a-3069"></a>

### #3069 — Is there a way to change the representation of one parameter of the request object?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3069)

**✅ Works-as-designed / 📖 Need doc (verified 2026-06-13).** codescan cannot
infer a custom `MarshalJSON`, but a `swagger:type` override (or
`encoding.TextMarshaler` / `swagger:strfmt`) states the representation
explicitly: with `swagger:type string` on the custom-marshalled `ComplexType`,
a `[]ComplexType` field renders as an array of strings — the reporter's actual
wire format. Locked by `internal/integration/coverage_bug_3069_test.go`
(fixture `fixtures/bugs/3069/`, golden `bugs_3069_schema.json`). 📖 Doc-site:
overriding the representation of custom-marshalled types.

<a id="a-3100"></a>

### #3100 — `in: formData` in `swagger:route` annotation translates to nothing (`in` field is omitted) in the yaml spec file

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3100)

**✅ Already fixed (verified 2026-06-12).** A parameter declared with
`in: formData` inside an inline `swagger:route` annotation now round-trips
its `in` field verbatim (the grammar YAML body parser preserves it);
confirmed for the reporter's `+name:` flush form, multiple `formData`
params, and a `query` param in the same block. Locked by
`internal/integration/coverage_bug_3100_test.go` (fixture
`fixtures/bugs/3100/`, golden `bugs_3100_schema.json`).

<a id="a-3107"></a>

### #3107 — No struct definition in swagger generate

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3107)

**✅ Already fixed (verified 2026-06-12).** Under `ScanModels` (= `generate
spec -m`), `MyStruct` is emitted with `type: object` and both properties —
no empty `{x-go-package}` stub. Verified across flat scan, recursive scan,
and the v0.30.5 trigger condition (model in its own package, resolved by
cross-package reference rather than direct scan; modern go/packages loading
resolves the type fully through the import graph). Locked by
`internal/integration/coverage_bug_3107_test.go` (fixture
`fixtures/bugs/3107/{api,model}`, golden `bugs_3107_schema.json`).


<a id="a-3117"></a>

### #3117 — Swagger spec generating type property along with schema references

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3117)

**✅ Fixed by grammar2 / lock (verified 2026-06-15).** A body parameter whose
schema is a `$ref` no longer carries an invalid sibling `type: object`; the
schema is a clean `$ref`. Locked by `fixtures/bugs/3117` + `TestCoverage_Bug3117`
(asserts the body schema is a bare `$ref` with empty `type`) + golden.


<a id="a-3119"></a>

### #3119 — Can not declare normal field in properties of schema object

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3119)

**✅ Works-as-designed / 📖 Need doc (verified 2026-06-13).** With consistent
YAML indentation, the scalar `limit`/`offset` fields land correctly inside
`properties` alongside the `data` array (the reporter's expected snippet 3).
The reported breakage came from mixed tabs/spaces in the comment YAML — a user
indentation mistake, not a parser bug. Cross-referenced on `fix/backlog-lot1`
(no fixture; the behaviour is already exercised by the inline
swagger:operation corpus). 📖 Doc-site: inline-schema YAML indentation must be
consistent (mixed tabs/spaces silently re-nests properties).


<a id="a-3134"></a>

### #3134 — How to Generate Swagger Specification for Versioned APIs in a Single Route File Using Go-Swagger?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3134)

**✅ Doc-only / covered + 📖 (re-triaged 2026-06-17).** The codescan-relevant part
of the question — how to annotate handlers and produce different spec files per
API version — is covered by the doc-site (`doc-site-update`):
`shaping-the-output/scoping-the-scan.md` explicitly documents "to produce several
specs from one module — e.g. one per API version — run a scan per package tree
(`./v1/...`, then `./v2/...`) and write each result separately," plus
Include/Exclude and tag filters; handler annotation is covered across
getting-started / annotation-index / tutorials. The "what command should I
execute" part is a go-swagger CLI concern (covered by go-swagger's own doc-site)
— out of scope here. The duplicate-`operationId` uniqueness diagnostic idea
spun off from this issue remains tracked as a forthcoming item (cross-linked from
forthcoming-features §14). 📖 Doc-site action: DONE (scoping-the-scan covers
per-version specs).


<a id="a-3138"></a>

### #3138 — How To mark a field as deprecated?

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3138)

**✅ Fixed / 📖 Need doc (2026-06-13).** Operation-level `deprecated: true`
keeps using the native OAS2 field. Model- and field-level deprecation (which
OAS2 `spec.Schema` can't express natively) now emits the `x-deprecated: true`
vendor extension.

Decision (with the maintainer): detection lives in the **grammar**, exposed as
`grammar.Block.IsDeprecated()`; builders consume that one signal. Two triggers,
on both models and fields:

- explicit `deprecated: true` annotation (consumed as the bool keyword;
  `deprecated: false` wins and stays not-deprecated);
- a godoc-style `Deprecated:` paragraph (pkgsite regexp), so the mark isn't
  repeated. The lexer leaves a `deprecated:` line whose arg is **not** a bool as
  prose (instead of forcing a bool parse), which (a) keeps the reason text in
  the description and (b) kills the spurious `parse.invalid-boolean` error the
  godoc form used to raise.

`x-deprecated` is emitted regardless of `SkipExtensions` (semantic intent, not
`x-go-*` reflection metadata). Implementation: `grammar/deprecated.go`
(`IsDeprecated()` + regexp), `lexer.go` (non-bool `deprecated:` → prose),
schema walker consumes `block.IsDeprecated()`, `resolvers.MarkDeprecated` /
`ExtDeprecated`. Verified by `internal/integration/coverage_bug_3138_test.go`
(golden `bugs_3138_schema.json`: explicit + godoc on field and model, reason
retained, op native field, no invalid-boolean diag) and grammar unit test
`TestBlock_IsDeprecated`. 📖 Doc-site: deprecated on operations is native; on
models/fields it rides `x-deprecated`, auto-detected from a godoc `Deprecated:`
paragraph.


<a id="a-3213"></a>

### #3213 — [spec/parsing] Consider TypeSpec comments

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3213)

**✅ Fixed (verified 2026-06-13).** Doc comments / annotations on an individual
TypeSpec inside a grouped `type ( ... )` declaration ARE parsed: two models in
one group keep their distinct titles (impossible for a GenDecl-only reader)
and a grouped `swagger:enum` has its const values parsed. Locked by
`internal/integration/coverage_bug_3213_test.go` (fixture `fixtures/bugs/3213/`,
golden `bugs_3213_schema.json`).


<a id="a-3214"></a>

### #3214 — [spec/parsing] Incomplete parsing of referenced typed primitives

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3214)

**✅ Already fixed (verified 2026-06-12, grammar2 migration).** The referenced
typed primitive `State` is emitted as its own definition with `enum`
correctly parsed and the prose preserved as `title`; the `enum:` declaration
line is never folded into the title/description — confirmed for both
`ScanModels` on and off. Locked by `internal/integration/coverage_bug_3214_test.go`
(fixture `fixtures/bugs/3214/`, golden `bugs_3214_schema.json`), which also
covers a multi-paragraph (`title` + `description`) variant (`Currency`).
