# Crear y poblar la base de datos

Ejecuta estos comandos desde la raíz del repositorio. Sustituye `tu_base_de_datos` por el nombre de la base de datos PostgreSQL que quieras usar:

```bash
psql -d tu_base_de_datos -f db/001_schema.sql
psql -d tu_base_de_datos -f db/002_seed.sql
psql -d tu_base_de_datos -f db/003_demo_seed.sql
```

El primer comando crea el esquema; el segundo carga catálogos (métricas, unidades y dispositivos); el tercero carga los tres astronautas, sus mediciones de tres días y la telemetría de cabina y traje.
