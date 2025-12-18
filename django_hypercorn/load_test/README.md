# Нагрузочное тестирование

## Определение максимальной производительности

Требует наличия/установки vegeta.

```bash
./benchmark.sh
```

Скрипт использует бинарный поиск для нахождения максимального RPS:
- Начинает с 1000 RPS
- При success ≥ 99.5% сохраняет как MIN_SUCCESS_RATE
  - Если нет известного MAX_FAIL_RATE → увеличивает на 50%
  - Если есть MAX_FAIL_RATE → берёт середину между MIN и MAX
- При success < 99.5% сохраняет как MAX_FAIL_RATE и берёт середину
- Останавливается когда разница между MIN и MAX ≤ 10 RPS или после 10 итераций

## Ручной запуск теста

```bash
vegeta attack -rate=100 -duration=10s -targets=vegeta_target.txt | vegeta report
```

## Параметры

- `-rate=100` - 100 запросов в секунду
- `-duration=10s` - длительность теста 10 секунд
- `-targets=vegeta_target.txt` - файл с описанием запроса

## Сохранение результатов в файл

```bash
vegeta attack -rate=100 -duration=10s -targets=vegeta_target.txt | tee results.bin | vegeta report
```

## Детальный отчет

```bash
vegeta attack -rate=100 -duration=10s -targets=vegeta_target.txt | vegeta report -type=text
```
