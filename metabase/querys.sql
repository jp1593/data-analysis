-- (a) Porcentaje de incidentes entre 2018-01-01 y 2020-12-31
SELECT 
  ROUND(
    (SUM(CASE WHEN date BETWEEN '2018-01-01' AND '2020-12-31' THEN 1 ELSE 0 END) / COUNT(*)) * 100,
    2
  ) AS Porcentaje_de_incidentes_entre_2018_y_2020
FROM incidents;

-- (b) Tres métodos de transporte más comunes detectados por “intelligence”
SELECT 
  transport_mode as Transport_mode,
  COUNT(*) AS Total
FROM details
WHERE detection LIKE '%intelligence%'
  AND transport_mode IS NOT NULL
  AND transport_mode != ''
GROUP BY transport_mode
ORDER BY total DESC
LIMIT 3;

-- (c) Métodos de detección con el mayor promedio de gente arrestada
SELECT 
  CASE
    WHEN LOWER(d.detection) LIKE '%intelligence%' THEN 'Intelligence'
    WHEN LOWER(d.detection) LIKE '%inspection%' THEN 'Routine Inspection'
    WHEN LOWER(d.detection) LIKE '%x-ray%' THEN 'X-ray'
    WHEN LOWER(d.detection) LIKE '%operation%' THEN 'Operation'
    WHEN LOWER(d.detection) LIKE '%investigation%' THEN 'Investigation'
    WHEN LOWER(d.detection) LIKE '%target%' THEN 'Targeting'
    WHEN LOWER(d.detection) LIKE '%risk%' THEN 'Risk Assessment'
    WHEN LOWER(d.detection) LIKE '%test%' THEN 'Test Purchase'
    WHEN LOWER(d.detection) LIKE '%dog%' THEN 'Dogs'
    WHEN LOWER(d.detection) LIKE '%online%' THEN 'Online'
    WHEN LOWER(d.detection) LIKE '%other%' THEN 'Other'
    WHEN LOWER(d.detection) LIKE '%drone%' THEN 'Drone'
    WHEN LOWER(d.detection) LIKE '%lake%' OR LOWER(d.detection) LIKE '%river%' THEN 'Lake/River'
    WHEN LOWER(d.detection) LIKE '%train%' THEN 'Land - Train'
    WHEN LOWER(d.detection) LIKE '%foot%' THEN 'Land - Foot'
    WHEN LOWER(d.detection) LIKE '%vehicle%' THEN 'Land - Vehicle'
    WHEN LOWER(d.detection) LIKE '%air%' THEN 'Air'
    WHEN LOWER(d.detection) LIKE '%sea%' THEN 'Sea'
    ELSE 'Other/Unclassified'
  END AS metodo_deteccion,
  ROUND(AVG(o.num_ppl_arrested), 2) AS promedio_arrestados
FROM details d
JOIN outcomes o ON d.report_id = o.report_id
WHERE d.detection IS NOT NULL
  AND d.detection != ''
GROUP BY metodo_deteccion
ORDER BY promedio_arrestados DESC;

-- (d) Categorías con las sentencias de prisión más largas
SELECT 
  i.category,
  ROUND(AVG(
    CASE 
      WHEN LOWER(TRIM(o.prison_time_unit)) LIKE 'year%' 
           AND TRIM(o.prison_time) REGEXP '^[0-9]+(\\.[0-9]+)?$'
        THEN CAST(TRIM(o.prison_time) AS DECIMAL(10,2)) * 365
      WHEN LOWER(TRIM(o.prison_time_unit)) LIKE 'month%' 
           AND TRIM(o.prison_time) REGEXP '^[0-9]+(\\.[0-9]+)?$'
        THEN CAST(TRIM(o.prison_time) AS DECIMAL(10,2)) * 30
      WHEN LOWER(TRIM(o.prison_time_unit)) LIKE 'day%' 
           AND TRIM(o.prison_time) REGEXP '^[0-9]+(\\.[0-9]+)?$'
        THEN CAST(TRIM(o.prison_time) AS DECIMAL(10,2))
      ELSE NULL
    END
  ), 2) AS promedio_dias_prision,
  
  ROUND(SUM(
    CASE 
      WHEN LOWER(TRIM(o.prison_time_unit)) LIKE 'year%' 
           AND TRIM(o.prison_time) REGEXP '^[0-9]+(\\.[0-9]+)?$'
        THEN CAST(TRIM(o.prison_time) AS DECIMAL(10,2)) * 365
      WHEN LOWER(TRIM(o.prison_time_unit)) LIKE 'month%' 
           AND TRIM(o.prison_time) REGEXP '^[0-9]+(\\.[0-9]+)?$'
        THEN CAST(TRIM(o.prison_time) AS DECIMAL(10,2)) * 30
      WHEN LOWER(TRIM(o.prison_time_unit)) LIKE 'day%' 
           AND TRIM(o.prison_time) REGEXP '^[0-9]+(\\.[0-9]+)?$'
        THEN CAST(TRIM(o.prison_time) AS DECIMAL(10,2))
      ELSE NULL
    END
  ), 2) AS total_dias_prision
FROM incidents i
JOIN outcomes o 
  ON i.report_id = o.report_id
WHERE o.prison_time IS NOT NULL
  AND TRIM(o.prison_time) <> ''
GROUP BY i.category
ORDER BY total_dias_prision DESC;

-- (e) Serie de tiempo anual con totales de multa por año
SELECT 
  YEAR(i.date) AS anio,
  SUM(o.fine) AS Total_multas
FROM incidents i
JOIN outcomes o ON i.report_id = o.report_id
GROUP BY anio
ORDER BY anio;
