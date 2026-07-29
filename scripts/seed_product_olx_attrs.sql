-- Preenche atributos OLX (tipo tela, resolução, Hz, condição, características)
-- com base nas specs de cada modelo. Idempotente por padrão de nome.

-- Helpers (features: curved, box, DP, HDR, wide, cables, audio, HDMI, ultrawide)

-- Dell E1709Wc 17" widescreen 1440x900 · VGA · LED
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%E1709W%' AND name NOT LIKE '%sem base%';

UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Aceitável',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%E1709W%' AND name LIKE '%sem base%';

-- Dell P1913Sb 19" 1280x1024 5:4 · VGA/DVI/DP
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1280x1024 (SXGA)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 1, feat_hdr = 0,
  feat_widescreen = 0, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%P1913%';

-- Dell P1914Sf 19" IPS 1280x1024 · VGA/DVI/DP · sem base
UPDATE products SET
  screen_type = 'IPS',
  max_resolution = '1280x1024 (SXGA)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Aceitável',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 1, feat_hdr = 0,
  feat_widescreen = 0, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%P1914%';

-- Dell P2016t 19,5" IPS 1440x900 · DP + VGA
UPDATE products SET
  screen_type = 'IPS',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 1, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%P2016%';

-- Dell P2219H 22" Full HD IPS · HDMI + DP + VGA · defeito na tela
UPDATE products SET
  screen_type = 'IPS',
  max_resolution = '1920x1080 (Full HD)',
  refresh_rate = '60 Hz',
  item_condition = 'Para peças / com defeito',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 1, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 1, feat_ultrawide = 0
WHERE name LIKE '%P2219%';

-- LG E1941S 18,5" 1366x768 TN · VGA · sem base
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1366x768 (HD)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Aceitável',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%E1941%';

-- LG W1942SE 19" 1440x900 · VGA · com base
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%W1942%';

-- LG W1943 (SC/SE) 1366x768 · VGA · com base
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1366x768 (HD)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%W1943%';

-- Lenovo ThinkVision L172 17" 1280x1024 · VGA · sem base
UPDATE products SET
  screen_type = 'LCD',
  max_resolution = '1280x1024 (SXGA)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Aceitável',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 0, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%L172%' OR name LIKE '%ThinkVision%';

-- Philips 236V4 23" Full HD · VGA + DVI · sem base
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1920x1080 (Full HD)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Aceitável',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%236V4%' OR name LIKE '%Philips 23%';

-- Prizi Slim 19" 1440x900 · HDMI + VGA
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 1, feat_ultrawide = 0
WHERE name LIKE '%Prizi%';

-- Samsung 733NW 17" 1440x900 · VGA
UPDATE products SET
  screen_type = 'LCD',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%733NW%' AND name NOT LIKE '%sem base%';

UPDATE products SET
  screen_type = 'LCD',
  max_resolution = '1440x900 (HD+)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Aceitável',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%733NW%' AND name LIKE '%sem base%';

-- Samsung 743B 17" 1280x1024 · VGA + DVI · com base
UPDATE products SET
  screen_type = 'LCD',
  max_resolution = '1280x1024 (SXGA)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 0, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%743B%';

-- Samsung B1630N 15,6" 1366x768 · VGA · com base
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1366x768 (HD)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%B1630%';

-- Samsung S19B300B 18,5" 1366x768 · VGA · com base
UPDATE products SET
  screen_type = 'LED',
  max_resolution = '1366x768 (HD)',
  refresh_rate = '60 Hz',
  item_condition = 'Usado - Bom',
  feat_curved = 0, feat_includes_box = 0, feat_displayport = 0, feat_hdr = 0,
  feat_widescreen = 1, feat_includes_cables = 0, feat_audio = 0, feat_hdmi = 0, feat_ultrawide = 0
WHERE name LIKE '%S19B300%';
