package notification

import (
	"crypto/tls"
	"fmt"
	"opensearch-alert/pkg/types"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	config *types.EmailConfig
	logger *logrus.Logger
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(config *types.EmailConfig, logger *logrus.Logger) *EmailNotifier {
	return &EmailNotifier{
		config: config,
		logger: logger,
	}
}

// IsEnabled 检查是否启用
func (e *EmailNotifier) IsEnabled() bool {
	return e.config.Enabled
}

// Send 发送邮件
func (e *EmailNotifier) Send(alert *types.Alert) error {
	if !e.IsEnabled() {
		return nil
	}

	e.logger.Debugf("开始发送邮件告警: %s (级别: %s)", alert.RuleName, alert.Level)

	// 验证邮件配置
	if err := e.validateConfig(); err != nil {
		e.logger.Errorf("邮件配置验证失败: %v", err)
		return fmt.Errorf("邮件配置错误: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.config.FromEmail)
	m.SetHeader("To", e.config.ToEmails...)
	m.SetHeader("Subject", fmt.Sprintf("[%s] %s", alert.Level, alert.RuleName))

	// 构建邮件内容
	body := e.buildEmailBody(alert)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.config.SMTPServer, e.config.SMTPPort, e.config.Username, e.config.Password)
	if e.config.UseTLS {
		d.TLSConfig = &tls.Config{ServerName: e.config.SMTPServer}
	}

	if err := d.DialAndSend(m); err != nil {
		e.logger.Errorf("邮件发送失败: %v", err)
		// 提供更详细的错误信息和建议
		if e.isQQMailError(err) {
			return fmt.Errorf("QQ邮箱认证失败，请检查授权码设置: %w", err)
		}
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	e.logger.Debugf("邮件消息发送成功，收件人: %v", e.config.ToEmails)
	e.logger.Infof("邮件告警已发送: %s", alert.RuleName)
	return nil
}

// buildEmailBody 构建邮件内容
func (e *EmailNotifier) buildEmailBody(alert *types.Alert) string {
	// 格式化告警消息，处理Markdown格式
	formattedMessage := e.formatMessageContent(alert.Message)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>KubeSphere-OpenSearch 告警通知</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            line-height: 1.6;
            color: #333;
        }
        .header { 
            background-color: #f8d7da; 
            border: 1px solid #f5c6cb; 
            padding: 20px; 
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .header h2 {
            margin: 0;
            color: #721c24;
        }
        .content { 
            margin: 20px 0; 
        }
        .field { 
            margin: 15px 0; 
            padding: 10px;
            background-color: #f8f9fa;
            border-left: 4px solid #007bff;
            border-radius: 4px;
        }
        .label { 
            font-weight: bold; 
            color: #495057;
            display: block;
            margin-bottom: 5px;
            font-size: 14px;
        }
        .value { 
            color: #212529;
            font-size: 16px;
            word-wrap: break-word;
        }
        .message-content {
            background-color: #e9ecef;
            padding: 15px;
            border-radius: 6px;
            margin: 10px 0;
            white-space: pre-wrap;
            font-family: 'Courier New', monospace;
        }
        .data { 
            background-color: #f8f9fa; 
            padding: 15px; 
            border-radius: 6px; 
            margin: 20px 0;
            border: 1px solid #dee2e6;
        }
        .data h4 {
            margin-top: 0;
            color: #495057;
        }
        .data pre {
            background-color: #ffffff;
            padding: 10px;
            border-radius: 4px;
            border: 1px solid #dee2e6;
            overflow-x: auto;
            font-size: 12px;
        }
        .level-critical { border-left-color: #dc3545; }
        .level-high { border-left-color: #fd7e14; }
        .level-medium { border-left-color: #ffc107; }
        .level-low { border-left-color: #28a745; }
        .level-info { border-left-color: #17a2b8; }
    </style>
</head>
<body>
    <div class="header">
        <h2>🚨 KubeSphere-OpenSearch 告警通知</h2>
    </div>
    
    <div class="content">
        <div class="field level-%s">
            <span class="label">规则名称:</span>
            <span class="value">%s</span>
        </div>
        <div class="field level-%s">
            <span class="label">告警级别:</span>
            <span class="value">%s</span>
        </div>
        <div class="field level-%s">
            <span class="label">触发时间:</span>
            <span class="value">%s</span>
        </div>
        <div class="field level-%s">
            <span class="label">匹配数量:</span>
            <span class="value">%d</span>
        </div>
        
        <div class="field level-%s">
            <span class="label">告警消息:</span>
            <div class="message-content">%s</div>
        </div>
        
        <div class="data">
            <h4>详细信息:</h4>
            <pre>%s</pre>
        </div>
    </div>
</body>
</html>
`, e.getLevelClass(alert.Level), alert.RuleName,
		e.getLevelClass(alert.Level), alert.Level,
		e.getLevelClass(alert.Level), alert.Timestamp.Format("2006-01-02 15:04:05"),
		e.getLevelClass(alert.Level), alert.Count,
		e.getLevelClass(alert.Level), formattedMessage,
		e.formatData(alert.Data))
}

// formatData 格式化数据
func (e *EmailNotifier) formatData(data map[string]interface{}) string {
	// 这里可以实现更复杂的数据格式化逻辑
	return fmt.Sprintf("%+v", data)
}

// formatMessageContent 格式化消息内容，处理Markdown格式
func (e *EmailNotifier) formatMessageContent(message string) string {
	// 将Markdown格式转换为HTML格式
	formatted := message

	// 处理代码块标记 ``` -> <pre><code>
	// 先处理代码块，避免与其他格式冲突
	formatted = e.processCodeBlocks(formatted)

	// 处理粗体标记 **text** -> <strong>text</strong>
	formatted = e.processBoldText(formatted)

	// 处理分隔线标记 --- -> <hr>
	formatted = strings.ReplaceAll(formatted, "---", "<hr>")

	// 处理换行符，确保正确显示
	formatted = strings.ReplaceAll(formatted, "\n", "<br>")

	return formatted
}

// processCodeBlocks 处理代码块
func (e *EmailNotifier) processCodeBlocks(text string) string {
	// 简单的代码块处理：``` -> <pre><code> 和 </code></pre>
	parts := strings.Split(text, "```")
	result := ""
	inCodeBlock := false

	for i, part := range parts {
		if i%2 == 0 {
			// 普通文本
			result += part
		} else {
			// 代码块内容
			if inCodeBlock {
				result += "</code></pre>" + part
				inCodeBlock = false
			} else {
				result += "<pre><code>" + part
				inCodeBlock = true
			}
		}
	}

	// 如果最后还在代码块中，需要关闭
	if inCodeBlock {
		result += "</code></pre>"
	}

	return result
}

// processBoldText 处理粗体文本
func (e *EmailNotifier) processBoldText(text string) string {
	// 处理粗体标记 **text** -> <strong>text</strong>
	parts := strings.Split(text, "**")
	result := ""
	inBold := false

	for i, part := range parts {
		if i%2 == 0 {
			// 普通文本
			result += part
		} else {
			// 粗体文本
			if inBold {
				result += "</strong>" + part
				inBold = false
			} else {
				result += "<strong>" + part
				inBold = true
			}
		}
	}

	// 如果最后还在粗体中，需要关闭
	if inBold {
		result += "</strong>"
	}

	return result
}

// getLevelClass 获取告警级别对应的CSS类名
func (e *EmailNotifier) getLevelClass(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	case "info":
		return "info"
	default:
		return "info"
	}
}

// validateConfig 验证邮件配置
func (e *EmailNotifier) validateConfig() error {
	if e.config.SMTPServer == "" {
		return fmt.Errorf("SMTP服务器地址不能为空")
	}
	if e.config.SMTPPort <= 0 {
		return fmt.Errorf("SMTP端口必须大于0")
	}
	if e.config.Username == "" {
		return fmt.Errorf("SMTP用户名不能为空")
	}
	if e.config.Password == "" {
		return fmt.Errorf("SMTP密码不能为空")
	}
	if e.config.FromEmail == "" {
		return fmt.Errorf("发件人邮箱不能为空")
	}
	if len(e.config.ToEmails) == 0 {
		return fmt.Errorf("收件人邮箱列表不能为空")
	}
	return nil
}

// isQQMailError 判断是否为QQ邮箱相关错误
func (e *EmailNotifier) isQQMailError(err error) bool {
	errStr := err.Error()
	return e.config.SMTPServer == "smtp.qq.com" &&
		(errStr == "535 Login fail. Account is abnormal, service is not open, password is incorrect, login frequency limited, or system is busy. More information at https://help.mail.qq.com/detail/108/1023" ||
			errStr == "535 Login fail. Account is abnormal, service is not open, password is incorrect, login frequency limited, or system is busy. More information at https://help.mail.qq.com/detail/108/1023")
}
