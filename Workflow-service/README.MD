Для генерации общего swagger-файла по отдельным схемам.

`swagger-cli bundle .\Workflow-service\internal\openapi\api.yaml --outfile generated/api.yaml --type yaml`

Из сгенерированного файла убрать index:

`C:\\Users\\simak\\go\\bin\\oapi-codegen.exe -config internal//openapi//config.yaml internal//openapi//generated//api.yaml`