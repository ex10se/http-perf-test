#!/bin/bash

DURATION=30s
MIN_SUCCESS_RATE=0
MAX_FAIL_RATE=999999

echo "=== Нагрузочное тестирование django+gunicorn+nginx ==="
echo "Длительность каждого теста: $DURATION"
echo "Бинарный поиск максимального RPS"
echo ""

RATE=1000
ITERATION=0
MAX_ITERATIONS=10

while [ $ITERATION -lt $MAX_ITERATIONS ]; do
    ITERATION=$((ITERATION + 1))
    
    echo "----------------------------------------"
    echo "Итерация $ITERATION: Тест $RATE RPS"
    echo "Диапазон поиска: [$MIN_SUCCESS_RATE, $MAX_FAIL_RATE]"
    echo "----------------------------------------"
    
    REPORT=$(vegeta attack -rate=$RATE -duration=$DURATION -targets=vegeta_target.txt | vegeta report -type=text)
    echo "$REPORT"
    
    SUCCESS=$(echo "$REPORT" | grep "Success" | grep -o '[0-9.]*%' | tr -d '%')
    
    echo ""
    echo "Success rate: $SUCCESS%"
    
    if (( $(awk -v s="$SUCCESS" 'BEGIN {print (s >= 99.5)}') )); then
        echo "✓ Успешный тест"
        MIN_SUCCESS_RATE=$RATE
        
        if [ $MAX_FAIL_RATE -eq 999999 ]; then
            NEW_RATE=$(awk -v r="$RATE" 'BEGIN {print int(r * 1.5)}')
        else
            NEW_RATE=$(awk -v min="$MIN_SUCCESS_RATE" -v max="$MAX_FAIL_RATE" 'BEGIN {print int((min + max) / 2)}')
        fi
    else
        echo "✗ Неуспешный тест"
        MAX_FAIL_RATE=$RATE
        NEW_RATE=$(awk -v min="$MIN_SUCCESS_RATE" -v max="$MAX_FAIL_RATE" 'BEGIN {print int((min + max) / 2)}')
    fi
    
    DIFF=$(awk -v max="$MAX_FAIL_RATE" -v min="$MIN_SUCCESS_RATE" 'BEGIN {print (max - min)}')
    
    if [ $NEW_RATE -eq $RATE ] || (( $(awk -v d="$DIFF" 'BEGIN {print (d <= 10 && d > 0)}') )); then
        echo ""
        echo "=== Достигнута точка сходимости ==="
        break
    fi
    
    RATE=$NEW_RATE
    echo ""
    sleep 2
done

echo ""
echo "=== Тестирование завершено ==="
echo "Максимальный стабильный RPS: ~$MIN_SUCCESS_RATE"
echo "Минимальный проблемный RPS: ~$MAX_FAIL_RATE"
