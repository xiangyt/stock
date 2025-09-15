#!/bin/bash

# 股东户数API测试脚本
# 用于测试东方财富股东户数API的响应

echo "=== 股东户数API测试 ==="

# 测试股票代码
STOCK_CODE="001208"  # 华菱线缆

echo "测试股票代码: $STOCK_CODE"
echo "API地址: https://datacenter-web.eastmoney.com/api/data/v1/get"

# 构建完整的API请求
API_URL="https://datacenter-web.eastmoney.com/api/data/v1/get"
PARAMS="callback=jQuery1123014159649525581786_1757941251037"
PARAMS="${PARAMS}&sortColumns=END_DATE"
PARAMS="${PARAMS}&sortTypes=-1"
PARAMS="${PARAMS}&pageSize=50"
PARAMS="${PARAMS}&pageNumber=1"
PARAMS="${PARAMS}&reportName=RPT_HOLDERNUM_DET"
PARAMS="${PARAMS}&columns=SECURITY_CODE%2CSECURITY_NAME_ABBR%2CCHANGE_SHARES%2CCHANGE_REASON%2CEND_DATE%2CINTERVAL_CHRATE%2CAVG_MARKET_CAP%2CAVG_HOLD_NUM%2CTOTAL_MARKET_CAP%2CTOTAL_A_SHARES%2CHOLD_NOTICE_DATE%2CHOLDER_NUM%2CPRE_HOLDER_NUM%2CHOLDER_NUM_CHANGE%2CHOLDER_NUM_RATIO%2CEND_DATE%2CPRE_END_DATE"
PARAMS="${PARAMS}&quoteColumns=f2%2Cf3"
PARAMS="${PARAMS}&quoteType=0"
PARAMS="${PARAMS}&filter=(SECURITY_CODE%3D%22${STOCK_CODE}%22)"
PARAMS="${PARAMS}&source=WEB"
PARAMS="${PARAMS}&client=WEB"

FULL_URL="${API_URL}?${PARAMS}"

echo ""
echo "发送请求..."
echo "URL: $FULL_URL"
echo ""

# 发送请求并保存响应
RESPONSE=$(curl -s \
  -H 'Accept: */*' \
  -H 'Accept-Language: zh-CN,zh;q=0.9,en;q=0.8' \
  -H 'Cache-Control: no-cache' \
  -H 'Connection: keep-alive' \
  -H 'Pragma: no-cache' \
  -H 'Referer: https://data.eastmoney.com/' \
  -H 'Sec-Fetch-Dest: script' \
  -H 'Sec-Fetch-Mode: no-cors' \
  -H 'Sec-Fetch-Site: same-site' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36' \
  "$FULL_URL")

echo "响应状态: $?"
echo ""
echo "响应内容:"
echo "$RESPONSE"
echo ""

# 尝试解析JSONP响应
if [[ $RESPONSE == *"jQuery"* ]]; then
    echo "检测到JSONP响应，尝试提取JSON数据..."
    
    # 提取JSON部分（去掉JSONP包装）
    JSON_DATA=$(echo "$RESPONSE" | sed 's/^[^(]*(//' | sed 's/);*$//')
    
    echo "提取的JSON数据:"
    echo "$JSON_DATA" | python3 -m json.tool 2>/dev/null || echo "$JSON_DATA"
    
    # 检查是否包含数据
    if [[ $JSON_DATA == *'"success":true'* ]]; then
        echo ""
        echo "✅ API请求成功"
        
        # 尝试提取记录数量
        COUNT=$(echo "$JSON_DATA" | grep -o '"count":[0-9]*' | cut -d':' -f2)
        if [[ -n $COUNT ]]; then
            echo "📊 返回记录数: $COUNT"
        fi
        
        # 检查是否有数据
        if [[ $JSON_DATA == *'"data":['* ]] && [[ $JSON_DATA != *'"data":[]'* ]]; then
            echo "📈 包含股东户数数据"
        else
            echo "⚠️  未找到股东户数数据"
        fi
    else
        echo ""
        echo "❌ API请求失败或返回错误"
    fi
else
    echo "❌ 响应格式异常，不是预期的JSONP格式"
fi

echo ""
echo "=== 测试完成 ==="