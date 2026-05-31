package codegen

// Language 编程语言类型
type Language string

const (
	LanguageCurl       Language = "curl"
	LanguagePython     Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageGo         Language = "go"
	LanguageJava       Language = "java"
	LanguagePHP        Language = "php"
)

// Template 代码模板
type Template struct {
	Language    Language
	Name        string
	Description string
	Generate    func(req *RequestData) string
}

// RequestData 请求数据
type RequestData struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        string
	BodyType    string
	Auth        *AuthData
}

// AuthData 认证数据
type AuthData struct {
	Type     string
	Username string
	Password string
	Token    string
	APIKey   string
	APIValue string
	Location string
}

// templates 模板注册表
var templates = map[Language]*Template{
	LanguageCurl: {
		Language:    LanguageCurl,
		Name:        "cURL",
		Description: "cURL 命令行",
		Generate:    generateCurl,
	},
	LanguagePython: {
		Language:    LanguagePython,
		Name:        "Python",
		Description: "Python requests 库",
		Generate:    generatePython,
	},
	LanguageJavaScript: {
		Language:    LanguageJavaScript,
		Name:        "JavaScript",
		Description: "JavaScript fetch API",
		Generate:    generateJavaScript,
	},
	LanguageGo: {
		Language:    LanguageGo,
		Name:        "Go",
		Description: "Go net/http",
		Generate:    generateGo,
	},
	LanguageJava: {
		Language:    LanguageJava,
		Name:        "Java",
		Description: "Java HttpURLConnection",
		Generate:    generateJava,
	},
	LanguagePHP: {
		Language:    LanguagePHP,
		Name:        "PHP",
		Description: "PHP cURL",
		Generate:    generatePHP,
	},
}

// generateCurl 生成 cURL 命令
func generateCurl(req *RequestData) string {
	code := "curl"

	// 方法
	if req.Method != "" && req.Method != "GET" {
		code += " -X " + req.Method
	}

	// 请求头
	for key, value := range req.Headers {
		code += " \\\n  -H '" + key + ": " + value + "'"
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			code += " \\\n  -u '" + req.Auth.Username + ":" + req.Auth.Password + "'"
		case "bearer":
			code += " \\\n  -H 'Authorization: Bearer " + req.Auth.Token + "'"
		case "apikey":
			if req.Auth.Location == "header" {
				code += " \\\n  -H '" + req.Auth.APIKey + ": " + req.Auth.APIValue + "'"
			}
		}
	}

	// 请求体
	if req.Body != "" {
		code += " \\\n  -d '" + req.Body + "'"
	}

	// URL
	code += " \\\n  '" + req.URL + "'"

	// 查询参数
	if len(req.QueryParams) > 0 {
		params := ""
		for key, value := range req.QueryParams {
			if params != "" {
				params += "&"
			}
			params += key + "=" + value
		}
		code += "?" + params
	}

	return code
}

// generatePython 生成 Python 代码
func generatePython(req *RequestData) string {
	code := "import requests\n\n"

	// URL
	code += "url = '" + req.URL + "'\n\n"

	// 查询参数
	if len(req.QueryParams) > 0 {
		code += "params = {\n"
		for key, value := range req.QueryParams {
			code += "    '" + key + "': '" + value + "',\n"
		}
		code += "}\n\n"
	}

	// 请求头
	if len(req.Headers) > 0 {
		code += "headers = {\n"
		for key, value := range req.Headers {
			code += "    '" + key + "': '" + value + "',\n"
		}
		code += "}\n\n"
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			code += "auth = ('" + req.Auth.Username + "', '" + req.Auth.Password + "')\n\n"
		case "bearer":
			if _, ok := req.Headers["Authorization"]; !ok {
				code += "headers['Authorization'] = 'Bearer " + req.Auth.Token + "'\n\n"
			}
		case "apikey":
			if req.Auth.Location == "header" {
				code += "headers['" + req.Auth.APIKey + "'] = '" + req.Auth.APIValue + "'\n\n"
			}
		}
	}

	// 请求体
	if req.Body != "" {
		if req.BodyType == "json" {
			code += "payload = " + req.Body + "\n\n"
		} else {
			code += "payload = '" + req.Body + "'\n\n"
		}
	}

	// 发送请求
	code += "response = requests." + req.Method + "(\n"
	code += "    url"
	if len(req.QueryParams) > 0 {
		code += ",\n    params=params"
	}
	if len(req.Headers) > 0 {
		code += ",\n    headers=headers"
	}
	if req.Auth != nil && req.Auth.Type == "basic" {
		code += ",\n    auth=auth"
	}
	if req.Body != "" {
		if req.BodyType == "json" {
			code += ",\n    json=payload"
		} else {
			code += ",\n    data=payload"
		}
	}
	code += "\n)\n\n"

	// 输出结果
	code += "print(response.status_code)\n"
	code += "print(response.text)\n"

	return code
}

// generateJavaScript 生成 JavaScript 代码
func generateJavaScript(req *RequestData) string {
	code := "// JavaScript Fetch API\n\n"

	// URL
	code += "const url = '" + req.URL + "';\n\n"

	// 请求头
	if len(req.Headers) > 0 {
		code += "const headers = {\n"
		for key, value := range req.Headers {
			code += "  '" + key + "': '" + value + "',\n"
		}
		code += "};\n\n"
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "bearer":
			code += "const token = '" + req.Auth.Token + "';\n\n"
		case "apikey":
			if req.Auth.Location == "header" {
				code += "const apiKey = '" + req.Auth.APIValue + "';\n\n"
			}
		}
	}

	// 请求配置
	code += "const options = {\n"
	code += "  method: '" + req.Method + "',\n"
	if len(req.Headers) > 0 {
		code += "  headers,\n"
	}
	if req.Body != "" {
		if req.BodyType == "json" {
			code += "  body: JSON.stringify(" + req.Body + "),\n"
		} else {
			code += "  body: '" + req.Body + "',\n"
		}
	}
	code += "};\n\n"

	// 发送请求
	code += "fetch(url, options)\n"
	code += "  .then(response => response.json())\n"
	code += "  .then(data => console.log(data))\n"
	code += "  .catch(error => console.error('Error:', error));\n"

	return code
}

// generateGo 生成 Go 代码
func generateGo(req *RequestData) string {
	code := "package main\n\n"
	code += "import (\n"
	code += "\t\"fmt\"\n"
	code += "\t\"io\"\n"
	code += "\t\"net/http\"\n"
	if req.Body != "" {
		code += "\t\"strings\"\n"
	}
	code += ")\n\n"
	code += "func main() {\n"

	// URL
	code += "\turl := \"" + req.URL + "\"\n\n"

	// 请求体
	if req.Body != "" {
		code += "\tpayload := strings.NewReader(`" + req.Body + "`)\n\n"
	}

	// 创建请求
	if req.Body != "" {
		code += "\treq, err := http.NewRequest(\"" + req.Method + "\", url, payload)\n"
	} else {
		code += "\treq, err := http.NewRequest(\"" + req.Method + "\", url, nil)\n"
	}
	code += "\tif err != nil {\n"
	code += "\t\tfmt.Println(\"Error:\", err)\n"
	code += "\t\treturn\n"
	code += "\t}\n\n"

	// 请求头
	for key, value := range req.Headers {
		code += "\treq.Header.Set(\"" + key + "\", \"" + value + "\")\n"
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			code += "\treq.SetBasicAuth(\"" + req.Auth.Username + "\", \"" + req.Auth.Password + "\")\n"
		case "bearer":
			code += "\treq.Header.Set(\"Authorization\", \"Bearer " + req.Auth.Token + "\")\n"
		case "apikey":
			if req.Auth.Location == "header" {
				code += "\treq.Header.Set(\"" + req.Auth.APIKey + "\", \"" + req.Auth.APIValue + "\")\n"
			}
		}
	}

	code += "\n"

	// 发送请求
	code += "\tclient := &http.Client{}\n"
	code += "\tresp, err := client.Do(req)\n"
	code += "\tif err != nil {\n"
	code += "\t\tfmt.Println(\"Error:\", err)\n"
	code += "\t\treturn\n"
	code += "\t}\n"
	code += "\tdefer resp.Body.Close()\n\n"

	// 读取响应
	code += "\tbody, err := io.ReadAll(resp.Body)\n"
	code += "\tif err != nil {\n"
	code += "\t\tfmt.Println(\"Error:\", err)\n"
	code += "\t\treturn\n"
	code += "\t}\n\n"

	code += "\tfmt.Println(\"Status:\", resp.StatusCode)\n"
	code += "\tfmt.Println(\"Body:\", string(body))\n"
	code += "}\n"

	return code
}

// generateJava 生成 Java 代码
func generateJava(req *RequestData) string {
	code := "import java.io.*;\n"
	code += "import java.net.*;\n"
	code += "import java.nio.charset.StandardCharsets;\n\n"
	code += "public class Main {\n"
	code += "    public static void main(String[] args) throws Exception {\n"

	// URL
	code += "        URL url = new URL(\"" + req.URL + "\");\n"
	code += "        HttpURLConnection conn = (HttpURLConnection) url.openConnection();\n\n"

	// 方法
	code += "        conn.setRequestMethod(\"" + req.Method + "\");\n"

	// 请求头
	for key, value := range req.Headers {
		code += "        conn.setRequestProperty(\"" + key + "\", \"" + value + "\");\n"
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			code += "        String auth = \"" + req.Auth.Username + ":" + req.Auth.Password + "\";\n"
			code += "        String encodedAuth = Base64.getEncoder().encodeToString(auth.getBytes());\n"
			code += "        conn.setRequestProperty(\"Authorization\", \"Basic \" + encodedAuth);\n"
		case "bearer":
			code += "        conn.setRequestProperty(\"Authorization\", \"Bearer " + req.Auth.Token + "\");\n"
		}
	}

	// 请求体
	if req.Body != "" {
		code += "\n        conn.setDoOutput(true);\n"
		code += "        try (OutputStream os = conn.getOutputStream()) {\n"
		code += "            byte[] input = \"" + req.Body + "\".getBytes(StandardCharsets.UTF_8);\n"
		code += "            os.write(input, 0, input.length);\n"
		code += "        }\n"
	}

	// 响应
	code += "\n        int responseCode = conn.getResponseCode();\n"
	code += "        System.out.println(\"Response Code: \" + responseCode);\n\n"

	code += "        BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream()));\n"
	code += "        String inputLine;\n"
	code += "        StringBuilder response = new StringBuilder();\n"
	code += "        while ((inputLine = in.readLine()) != null) {\n"
	code += "            response.append(inputLine);\n"
	code += "        }\n"
	code += "        in.close();\n\n"

	code += "        System.out.println(response.toString());\n"
	code += "    }\n"
	code += "}\n"

	return code
}

// generatePHP 生成 PHP 代码
func generatePHP(req *RequestData) string {
	code := "<?php\n\n"

	// URL
	code += "$url = '" + req.URL + "';\n\n"

	// 初始化 cURL
	code += "$ch = curl_init();\n\n"

	// 设置选项
	code += "curl_setopt($ch, CURLOPT_URL, $url);\n"
	code += "curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);\n"

	// 方法
	if req.Method == "POST" {
		code += "curl_setopt($ch, CURLOPT_POST, true);\n"
	} else if req.Method != "GET" {
		code += "curl_setopt($ch, CURLOPT_CUSTOMREQUEST, '" + req.Method + "');\n"
	}

	// 请求头
	if len(req.Headers) > 0 {
		code += "\n$headers = [\n"
		for key, value := range req.Headers {
			code += "    '" + key + ": " + value + "',\n"
		}
		code += "];\n"
		code += "curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);\n"
	}

	// 请求体
	if req.Body != "" {
		code += "\n$data = '" + req.Body + "';\n"
		code += "curl_setopt($ch, CURLOPT_POSTFIELDS, $data);\n"
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			code += "\ncurl_setopt($ch, CURLOPT_USERPWD, '" + req.Auth.Username + ":" + req.Auth.Password + "');\n"
		case "bearer":
			code += "\n$headers[] = 'Authorization: Bearer " + req.Auth.Token + "';\n"
			code += "curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);\n"
		}
	}

	// 执行请求
	code += "\n$response = curl_exec($ch);\n"
	code += "$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);\n\n"

	code += "if (curl_errno($ch)) {\n"
	code += "    echo 'Error: ' . curl_error($ch);\n"
	code += "} else {\n"
	code += "    echo 'Status Code: ' . $httpCode . \"\\n\";\n"
	code += "    echo $response;\n"
	code += "}\n\n"

	code += "curl_close($ch);\n"
	code += "?>\n"

	return code
}
