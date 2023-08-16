Конечно, вот пример `README.md` для кода экспортера OOM:

```markdown
# OOM Exporter

OOM Exporter - это инструмент для мониторинга событий Out-Of-Memory (OOM) в системных логах и экспортирования метрик в формате, понимаемом Prometheus.

## Описание

OOM Exporter отслеживает события OOM в системных логах и генерирует метрики Prometheus для последующего мониторинга и анализа. Этот инструмент решает проблему отслеживания событий исчерпания памяти в реальном времени и предоставляет метрики для агрегированной статистики.

## Установка и использование

1. Склонируйте репозиторий:

```sh
git clone https://github.com/yourusername/oom_exporter.git
```

2. Перейдите в каталог проекта:

```sh
cd oom_exporter
```

3. Отредактируйте `config.yaml` для указания параметров мониторинга и экспорта:

```yaml
log_file: /var/log/syslog           # Путь к файлу логов
log_pattern: "Out of memory"        # Шаблон для поиска событий OOM
exporter_port: "9090"               # Порт для экспорта метрик
repeat_interval: 60                 # Интервал повторения мониторинга (секунды)
state_file: state.yaml             # Файл состояния для сохранения позиции чтения
```

4. Соберите и запустите экспортер:

```sh
go build
./oom_exporter
```

5. Подключитесь к Prometheus:

Добавьте следующую конфигурацию в ваш файл `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'oom_exporter'
    static_configs:
      - targets: ['localhost:9090']  # Адрес, на котором работает экспортер
```

6. Перезапустите Prometheus:

```sh
sudo systemctl restart prometheus
```

7. После этого, метрики OOM будут доступны по адресу `http://localhost:9090/metrics`.

## Лицензия

Этот проект распространяется под лицензией [MIT License](LICENSE).

```

Пожалуйста, убедитесь, что вы адаптировали пути и настройки в `config.yaml` и `prometheus.yml` под свою конфигурацию и потребности.
