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
    this.typeMapping.put("File", "*osFile");
    this.typeMapping.put("file", "*osFile");
    this.typeMapping.put("binary", "*osFile");
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
  }
}

CollectionCodegen.main(args)