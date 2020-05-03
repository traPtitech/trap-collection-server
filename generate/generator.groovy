@Grab(group = 'org.openapitools', module = 'openapi-generator-cli', version = '4.0.0')

import org.openapitools.codegen.*
import org.openapitools.codegen.languages.*

class CollectionCodegen extends GoGinServerCodegen {
  static main(String[] args) {
    OpenAPIGenerator.main(args)
  }

  CollectionCodegen() {
    super()
    this.apiPath = "openapi"
    this.typeMapping.put("File", "ioReader");
    this.typeMapping.put("file", "ioReader");
    this.typeMapping.put("binary", "ioReader");
  }

  @Override
  public String apiPackage() {
    return apiPath
  }

  @Override
  public void processOpts() {
    super.processOpts()

    if (additionalProperties.containsKey("apiPath")) {
      this.apiPath = (String)additionalProperties.get("apiPath");
    } else {
      additionalProperties.put("apiPath", apiPath);
    }

    modelPackage = packageName;
    apiPackage = packageName;

    this.supportingFiles = new ArrayList<SupportingFile>();
    supportingFiles.add(new SupportingFile("main.mustache", "", "main.go"))
    writeOptional(outputFolder, new SupportingFile("go.mod", "", "go.mod"))
    supportingFiles.add(new SupportingFile("router.mustache", apiPath, "router.go"))
    supportingFiles.add(new SupportingFile("interfaces.mustache", apiPath, "interfaces.go"))
    supportingFiles.add(new SupportingFile("README.mustache", "", "README.md"))
    supportingFiles.add(new SupportingFile("mockgen.mustache", "", "mockgen.sh"))
  }

  @Override
  public Map<String, Object> postProcessOperationsWithModels(Map<String, Object> operations, List<Object> allModels) {
    super.postProcessOperationsWithModels(operations, allModels)

    Map<String, Object> objs = (Map<String, Object>) operations.get("operations");
    List<CodegenOperation> ops = (List<CodegenOperation>) objs.get("operation");
    for (CodegenOperation op : ops) {
      op.httpMethod = op.httpMethod.toUpperCase(Locale.ENGLISH)

      op.baseName = op.nickname

      List<String> authMethodNames = new ArrayList<String>()
      List<CodegenSecurity> authMethods = new ArrayList<CodegenSecurity>()
      List<CodegenSecurity> otherAuthMethods = new ArrayList<CodegenSecurity>()
      for (CodegenSecurity authMethod : op.authMethods) {
        if (authMethod.name!="LauncherAuth" && authMethod.name!="TrapMemberAuth"){
          authMethods.add(authMethod)
        } else {
          authMethodNames.add(authMethod.name)
          otherAuthMethods.add(authMethod)
        }
      }
      if (authMethodNames == new ArrayList<String>(Arrays.asList("LauncherAuth", "TrapMemberAuth"))) {
        CodegenSecurity authMethod = new CodegenSecurity()
        authMethod.name = "BothAuth"
        otherAuthMethods = new ArrayList<CodegenSecurity>(Arrays.asList(authMethod))
      }
      otherAuthMethods.addAll(authMethods)
      op.authMethods = otherAuthMethods

      List<CodegenParameter> cookieParams = new ArrayList<CodegenParameter>()
      for (CodegenParameter cookieParam : op.cookieParams) {
        if (cookieParam.paramName == "sessions") {
          cookieParam.isCookieParam = false
        }
        cookieParams.add(cookieParam)
      }
      op.cookieParams = cookieParams

      List<CodegenParameter> allParams = new ArrayList<CodegenParameter>()
      for (CodegenParameter allParam : op.allParams) {
        if (allParam.paramName == "sessions") {
          allParam.isCookieParam = false
        }
        allParams.add(allParam)
      }
      op.allParams = allParams
    }
    return operations;
  }
}

CollectionCodegen.main(args)