-- Technical specs + marketplace listing copy (idempotent updates by name pattern)

-- Dell E1709Wc
UPDATE products SET
  description = 'Tela 17" widescreen · Resolução nativa 1440×900 (16:10) · Brilho ~250 cd/m² · Tempo de resposta ~8 ms · Entrada principal: VGA · Monitor office entry-level Dell série E.',
  listing_text = 'Monitor Dell E1709W 17" widescreen 1440x900. Ideal para uso básico, planilhas e navegação. Conexão VGA. Usado, testado. Entrega/retirada a combinar. Fotos reais do lote.'
WHERE name LIKE '%E1709W%';

-- Dell P1913Sb
UPDATE products SET
  description = 'Tela 19" · Resolução 1280×1024 (proporção 5:4 “quadrado”) · Série Dell Professional P1913S · Entradas típicas: VGA, DVI e DisplayPort (varia por revisão) · Bom para documentos e sistemas legados.',
  listing_text = 'Monitor Dell P1913S 19" 1280x1024 (5:4). Formato quadrado excelente para escritório e sistemas antigos. Usado, testado. Verificar cabos (VGA/DVI/DP).'
WHERE name LIKE '%P1913%';

-- Dell P1914Sf
UPDATE products SET
  description = 'Tela 19" IPS · Resolução 1280×1024 (5:4) · Ângulos ~178° · Brilho ~250 cd/m² · Contraste típ. 1000:1 · Entradas: VGA, DVI, DisplayPort · Painel IPS (melhor ângulo que TN).',
  listing_text = 'Monitor Dell P1914S 19" IPS 1280x1024 (5:4). Cores e ângulo melhores que TN. Usado, testado. Sem base em algumas unidades — VESA/suporte a combinar.'
WHERE name LIKE '%P1914%';

-- Dell P2016t
UPDATE products SET
  description = 'Tela 19,5" IPS W-LED · Resolução 1440×900 (16:10) · Ângulos 178°/178° · Brilho 250 cd/m² · Contraste 1000:1 · Resposta ~6–8 ms · Entradas: DisplayPort 1.2, VGA, hub USB 2.0 · Ergonomia (altura/pivot/swivel) quando com base completa.',
  listing_text = 'Monitor Dell P2016 19,5" IPS 1440x900. Painel IPS, DisplayPort + VGA. Linha Professional. Usado, testado. Fotos do produto.'
WHERE name LIKE '%P2016%';

-- Dell P2219H (incl. defeito)
UPDATE products SET
  description = 'Tela 21,5" viewable (rotulada 22") · Full HD 1920×1080 @ 60 Hz · Painel IPS · Brilho ~250 cd/m² · Entradas: HDMI 1.4, DisplayPort 1.2, VGA · Hub USB · Série Dell P. Unidade com defeito na parte inferior da imagem — ver fotos/vídeo.',
  listing_text = 'Monitor Dell P2219H 21,5"/22" Full HD IPS HDMI/DP/VGA. ATENÇÃO: defeito na parte inferior da tela (ver fotos e vídeo). Ideal se o preço compensar. Usado, testado. Preço já reflete o defeito.'
WHERE name LIKE '%P2219%';

-- LG E1941S
UPDATE products SET
  description = 'Tela 18,5" LED · Resolução 1366×768 (16:9) · Painel TN · Brilho 250 cd/m² · Resposta 5 ms · Entrada VGA · Série LG E41 entry LED.',
  listing_text = 'Monitor LG E1941S 18,5" LED 1366x768. Leve e econômico. VGA. Usado, testado. Sem base em algumas unidades.'
WHERE name LIKE '%E1941%';

-- LG W1942SE
UPDATE products SET
  description = 'Tela ~19" wide · Resolução 1440×900 (16:10) · Painel TN · Brilho ~300 cd/m² · Resposta ~5 ms · Entrada VGA · Flatron wide.',
  listing_text = 'Monitor LG W1942 19" widescreen 1440x900. Uso geral. VGA. Usado, com base, testado.'
WHERE name LIKE '%W1942%';

-- LG W1943SC / W1943SE
UPDATE products SET
  description = 'Tela 18,5–19" wide LED · Resolução 1366×768 (16:9) · Painel TN · Brilho 250 cd/m² · Resposta 5 ms · Contraste dinâmico alto · Entrada VGA · Família Flatron W1943.',
  listing_text = 'Monitor LG W1943 18,5"/19" LED HD 1366x768. Com base. Usado, testado. (Variantes SC/SE; moldura HP quando indicado é apenas cosmético.)'
WHERE name LIKE '%W1943%';

-- Lenovo L172
UPDATE products SET
  description = 'Tela 17" · Resolução 1280×1024 (5:4) · ThinkVision L172 (P/N 4428-AB1) · Entrada VGA · Monitor corporativo Lenovo · VESA em muitas unidades.',
  listing_text = 'Monitor Lenovo ThinkVision L172 17" 1280x1024. Formato 5:4 para escritório. VGA. Usado, sem base em algumas unidades, testado.'
WHERE name LIKE '%L172%' OR name LIKE '%ThinkVision%';

-- Philips 236V4
UPDATE products SET
  description = 'Tela 23" LED · Full HD 1920×1080 @ 60 Hz · Brilho 250 cd/m² · Resposta ~5 ms · Entradas: VGA + DVI-D (HDCP) · Aspecto 16:9 · Boa opção de tamanho no lote.',
  listing_text = 'Monitor Philips 23" Full HD 1920x1080 LED. VGA e DVI. Usado, sem base em algumas unidades, testado. Ótimo custo por polegada.'
WHERE name LIKE '%236V4%' OR name LIKE '%Philips 23%';

-- Prizi
UPDATE products SET
  description = 'Tela 19" LED Slim · Resolução 1440×900 (16:10) · Brilho ~250 cd/m² · Entradas: HDMI + VGA · Marca Prizi (mercado BR) · Modelo PZ0018HDMI / linha Slim 19".',
  listing_text = 'Monitor Prizi Slim 19" LED 1440x900 HDMI e VGA. Fácil de ligar em notebook/TV box via HDMI. Usado, testado. Fotos reais.'
WHERE name LIKE '%Prizi%';

-- Samsung 733NW
UPDATE products SET
  description = 'Tela 17" widescreen · Resolução 1440×900 · Brilho 250 cd/m² · Resposta ~8 ms · Contraste típ. 600–1000:1 (DC alto) · Entrada VGA · SyncMaster 733NW.',
  listing_text = 'Monitor Samsung SyncMaster 733NW 17" 1440x900. Widescreen clássico. VGA. Usado, testado. Com ou sem base conforme anúncio.'
WHERE name LIKE '%733NW%';

-- Samsung 743B
UPDATE products SET
  description = 'Tela 17" · Resolução 1280×1024 (5:4) · Brilho ~300 cd/m² · Resposta 5 ms · Contraste ~1000:1 (DC ~7000:1) · Entradas VGA + DVI · SyncMaster 743B.',
  listing_text = 'Monitor Samsung SyncMaster 743B 17" 1280x1024. Formato 5:4. VGA/DVI. Usado, com base, testado.'
WHERE name LIKE '%743B%';

-- Samsung B1630N
UPDATE products SET
  description = 'Tela 15,6" wide · Resolução 1360/1366×768 · Brilho 250–300 cd/m² · Resposta 5–8 ms · Entrada VGA · Compacto, ângulo de visão limitado (TN).',
  listing_text = 'Monitor Samsung B1630N 15,6" 1366x768. Compacto para espaços pequenos. VGA. Usado, com base, testado.'
WHERE name LIKE '%B1630%';

-- Samsung S19B300B
UPDATE products SET
  description = 'Tela 18,5" LED · Resolução 1366×768 (16:9) · Brilho 250 cd/m² · Resposta 5 ms · Contraste ~1000:1 / Mega DCR · Entrada VGA · Série B300 slim.',
  listing_text = 'Monitor Samsung S19B300B 18,5" LED HD 1366x768. Design slim. VGA. Usado, com base, testado.'
WHERE name LIKE '%S19B300%';
