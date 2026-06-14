# 09 LLM Integration via OpenRouter

> Spesifikasi teknis untuk memanggil LLM via OpenRouter API.

---

## 9.1 Why OpenRouter?

| Criteria | OpenRouter | AWS Bedrock (production) |
|----------|-----------|---------------------------|
| API Key Setup | Instant, free trial available | AWS account + IAM + region |
| Model Access | Claude, GPT-4o, Llama, etc. | Claude, Titan, Llama (region-limited) |
| Billing | Prepaid / trial | AWS billing complex |
| Latency | Fast (edge distributed) | Regional, occasional throttling |
| Switch Cost to Bedrock | Low (sama format API) | N/A |

**Strategy:** Start OpenRouter untuk demo. Nantinya switch ke AWS Bedrock dengan mengganti base URL dan API key saja.

---

## 9.2 API Specification

### Base URL

```
https://openrouter.ai/api/v1/chat/completions
```

### Headers

```
Authorization: Bearer {OPENROUTER_API_KEY}
Content-Type: application/json
HTTP-Referer: https://company.com (opsional)
X-Title: AI Incident Commander (opsional)
```

### Request Body

```json
{
  "model": "anthropic/claude-sonnet-4-20250514",
  "messages": [
    {
      "role": "system",
      "content": "You are the Infrastructure & Platform Engineering AI..."
    },
    {
      "role": "user",
      "content": "Service: payment-api\nIncident Type: outage\n...\nAnalyze and return JSON: {\"finding\":\"...\",\"confidence\":92,\"suggested_action\":\"...\",\"evidence\":[\"...\"]}"
    }
  ],
  "temperature": 0.2,
  "max_tokens": 1024,
  "response_format": {
    "type": "json_object"
  }
}
```

### Response Format

```json
{
  "id": "gen-1234567890",
  "model": "anthropic/claude-sonnet-4-20250514",
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "{\"finding\":\"...\",\"confidence\":92,\"suggested_action\":\"...\",\"evidence\":[\"...\"]}"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ],
  "usage": {
    "prompt_tokens": 1250,
    "completion_tokens": 342,
    "total_tokens": 1592,
    "total_cost": 0.0042
  }
}
```

---

## 9.3 JavaScript/Node.js Implementation in Python (using httpx)

```python
# app/agents/base.py (extract)

import httpx
import json
import asyncio
from typing import Dict, Any, Optional
from app.config import settings

class LLMClient:
    def __init__(self):
        self.base_url = "https://openrouter.ai/api/v1"
        self.api_key = settings.OPENROUTER_API_KEY
        self.model = settings.LLM_MODEL
        self.client = httpx.AsyncClient(
            timeout=30.0,
            headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
                "HTTP-Referer": "https://company.com",
                "X-Title": "AI Incident Commander"
            }
        )
    
    async def call(
        self,
        system_prompt: str,
        user_prompt: str,
        temperature: float = 0.2,
        max_tokens: int = 1024,
        response_format_type: str = "json_object"
    ) -> Dict[str, Any]:
        """
        Call OpenRouter API and return parsed JSON response.
        """
        payload = {
            "model": self.model,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt}
            ],
            "temperature": temperature,
            "max_tokens": max_tokens,
            "response_format": {"type": response_format_type}
        }
        
        try:
            response = await self.client.post(
                f"{self.base_url}/chat/completions",
                json=payload
            )
            response.raise_for_status()
            data = response.json()
            
            # Extract content
            content = data["choices"][0]["message"]["content"]
            
            # Parse JSON
            result = json.loads(content)
            
            # Extract token usage
            usage = data.get("usage", {})
            
            return {
                "success": True,
                "data": result,
                "raw_response": content,
                "tokens_prompt": usage.get("prompt_tokens", 0),
                "tokens_completion": usage.get("completion_tokens", 0),
                "tokens_total": usage.get("total_tokens", 0),
                "cost_usd": usage.get("total_cost", 0.0),
                "model": data.get("model", self.model),
                "latency_ms": response.elapsed.total_seconds() * 1000
            }
            
        except httpx.TimeoutException:
            return {
                "success": False,
                "error": "LLM_TIMEOUT",
                "error_detail": "OpenRouter API timeout after 30s"
            }
        except json.JSONDecodeError as e:
            return {
                "success": False,
                "error": "INVALID_JSON",
                "error_detail": f"Failed to parse LLM response: {str(e)}",
                "raw_response": content if 'content' in locals() else None
            }
        except httpx.HTTPStatusError as e:
            return {
                "success": False,
                "error": "HTTP_ERROR",
                "error_detail": f"HTTP {e.response.status_code}: {e.response.text}"
            }
        except Exception as e:
            return {
                "success": False,
                "error": "UNKNOWN_ERROR",
                "error_detail": str(e)
            }
    
    async def close(self):
        await self.client.aclose()
```

---

## 9.4 Retry Strategy

```python
async def call_with_retry(self, system_prompt, user_prompt, max_retries=2):
    for attempt in range(max_retries + 1):
        result = await self.call(system_prompt, user_prompt)
        if result["success"]:
            return result
        if result["error"] in ["INVALID_JSON"] and attempt < max_retries:
            # Retry with stricter prompt
            user_prompt += "\n\nIMPORTANT: Respond ONLY with valid JSON."
            continue
        if attempt == max_retries:
            return result
    return result
```

---

## 9.5 Switch to AWS Bedrock (Future)

Untuk switch ke AWS Bedrock:

1. Ganti `base_url` ke AWS endpoint:
   ```python
   # Production: AWS Bedrock
   base_url = "https://bedrock-runtime.{region}.amazonaws.com"
   ```

2. Ganti `Authorization` ke AWS Signature V4:
   ```python
   headers = {
       "Content-Type": "application/json",
       "Accept": "application/json"
   }
   # Use boto3 for SigV4 signing
   ```

3. Atau via Amazon Bedrock Invoke API:
   ```python
   import boto3
   runtime = boto3.client('bedrock-runtime', region_name='us-east-1')
   response = runtime.invoke_model(
       modelId="anthropic.claude-sonnet-4-20250514-v1",
       body=json.dumps({"prompt": system_prompt + "\n\n" + user_prompt})
   )
   ```

**Perbedaan yang harus diadaptasi:**
- Bedrock tidak native support `response_format: {type: "json_object"}` — gunakan instruction engineering
- Token counting beda (Bedrock menggunakan format prompt Anthropic)
- Error codes beda

---

## 9.6 Cost Estimation (Demo)

| Call Type | Est Tokens | Est Cost/Call |
|-----------|-----------|---------------|
| Infrastructure Agent | ~1,500 prompt + ~300 completion | ~$0.004 |
| Application Agent | ~1,500 prompt + ~300 completion | ~$0.004 |
| Change Correlation | ~1,200 prompt + ~200 completion | ~$0.003 |
| Business Impact | ~1,000 prompt + ~200 completion | ~$0.003 |
| Decision Engine | ~2,500 prompt + ~400 completion | ~$0.007 |
| Total per incident | ~7,700 tokens | ~$0.021 |

**Demo Budget:**
- 20 trigger test × $0.021 = **$0.42**
- Sangat aman untuk demo.

---

## 9.7 Prompt Engineering Checklist

- [ ] System prompt ≤ 500 tokens
- [ ] User prompt jelas, berisi semua context
- [ ] JSON schema didefinisikan di prompt
- [ ] Temperature ~0.2 (deterministik)
- [ ] Max tokens cukup (512–1024)
- [ ] Fallback jika JSON parse fail

---

> Next: baca `10-approval-workflow.md` untuk approval flow detail.
