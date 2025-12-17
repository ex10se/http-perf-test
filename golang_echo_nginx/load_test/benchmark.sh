#!/bin/bash

set -e

DURATION=30s
MIN_SUCCESS_RATE=0
MAX_FAIL_RATE=999999
THRESHOLD=99.5
MAX_RATE_LIMIT=100000

echo "=== Нагрузочное тестирование golang+gin+nginx ==="
echo "Длительность: $DURATION, Цель: >= $THRESHOLD%"

RATE=1000
ITERATION=0
MAX_ITERATIONS=20
FOUND_UPPER_BOUND=0
BEST_REPORT=""

while [ $ITERATION -lt $MAX_ITERATIONS ]; do
    ITERATION=$((ITERATION + 1))

    echo "----------------------------------------"
    echo "Итерация $ITERATION: $RATE RPS"
    if [ $MAX_FAIL_RATE -eq 999999 ]; then
        echo "Режим: Экспоненциальный рост (ищем верхний предел)"
    else
        echo "Режим: Бинарный поиск (точная настройка)"
    fi
    echo "Диапазон: [$MIN_SUCCESS_RATE, $MAX_FAIL_RATE]"
    echo "----------------------------------------"

    if [ $RATE -gt $MAX_RATE_LIMIT ]; then
        echo "⚠️ Достигнут лимит $MAX_RATE_LIMIT RPS, останавливаемся"
        MAX_FAIL_RATE=$RATE
        FOUND_UPPER_BOUND=1
        break
    fi

    REPORT=$(timeout 120 vegeta attack -rate=$RATE -duration=$DURATION \
        -targets=vegeta_target.txt 2>&1 | timeout 60 vegeta report -type=text) || {
        echo "❌ Ошибка vegeta"
        exit 1
    }

    echo "$REPORT"
    SUCCESS=$(echo "$REPORT" | awk '/^Requests.*\[/ {next} /Success/ {match($0, /[0-9]+\.[0-9]+%/); print substr($0, RSTART, RLENGTH-1)}')

    if [ -z "$SUCCESS" ]; then
        echo "❌ Ошибка парсинга"
        exit 1
    fi

    echo ""
    echo "Success rate: $SUCCESS%"
    echo ""

    if (( $(awk -v s="$SUCCESS" -v t="$THRESHOLD" 'BEGIN {print (s >= t)}') )); then
        echo "✓ Успешный тест"
        MIN_SUCCESS_RATE=$RATE
        BEST_REPORT="$REPORT"

        if [ $MAX_FAIL_RATE -eq 999999 ]; then
            NEW_RATE=$(awk -v r="$RATE" 'BEGIN {
                inc = int(r * 0.5)
                if (inc < 500) inc = 500
                print r + inc
            }')
        else
            NEW_RATE=$(awk -v min="$MIN_SUCCESS_RATE" -v max="$MAX_FAIL_RATE" 'BEGIN {print int((min + max) / 2)}')
        fi
    else
        echo "✗ Неуспешный тест"
        MAX_FAIL_RATE=$RATE
        FOUND_UPPER_BOUND=1

        if [ $MIN_SUCCESS_RATE -eq 0 ]; then
            echo "⚠️ Первый тест упал, снижаем RPS"
            NEW_RATE=$(awk -v r="$RATE" 'BEGIN {print int(r / 2)}')
        else
            NEW_RATE=$(awk -v min="$MIN_SUCCESS_RATE" -v max="$MAX_FAIL_RATE" 'BEGIN {print int((min + max) / 2)}')
        fi
    fi

    DIFF=$(awk -v max="$MAX_FAIL_RATE" -v min="$MIN_SUCCESS_RATE" 'BEGIN {print (max - min)}')

    if [ $FOUND_UPPER_BOUND -eq 1 ] && (( $(awk -v d="$DIFF" 'BEGIN {print (d <= 5)}') )); then
        echo "=== Сходимость достигнута (диапазон < 5 RPS) ==="
        break
    fi

    if [ $NEW_RATE -eq $RATE ]; then
        echo "=== Сходимость достигнута (RPS не изменился) ==="
        break
    fi

    RATE=$NEW_RATE
    echo ""
    sleep 2
done

echo ""
echo "=== Результаты ==="
if [ $MIN_SUCCESS_RATE -eq 0 ]; then
    echo "❌ Система не справляется даже с минимальной нагрузкой"
else
    echo "Максимальный стабильный RPS: $MIN_SUCCESS_RATE"
    echo "Минимальный проблемный RPS: $MAX_FAIL_RATE"

    if [ $FOUND_UPPER_BOUND -eq 1 ]; then
        MIDPOINT=$(awk -v min="$MIN_SUCCESS_RATE" -v max="$MAX_FAIL_RATE" 'BEGIN {print int((min + max) / 2)}')
        SAFE_RATE=$(awk -v m="$MIDPOINT" 'BEGIN {print int(m * 0.8)}')
        echo "Рекомендуемый RPS в продакшене: $SAFE_RATE"
    else
        echo "⚠️ Верхний предел не найден (система держит > $MIN_SUCCESS_RATE RPS)"
    fi

    echo ""
    echo "=== Детальный отчёт vegeta для максимального стабильного RPS ($MIN_SUCCESS_RATE) ==="
    echo "$BEST_REPORT"
fi
