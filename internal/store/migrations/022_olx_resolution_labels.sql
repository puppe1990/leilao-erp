-- Align products.max_resolution with OLX category labels.
UPDATE products SET max_resolution = 'até 720p - HD'
WHERE max_resolution IN (
  '1280x1024 (SXGA)',
  '1366x768 (HD)',
  '1440x900 (HD+)',
  '1600x900 (HD+)',
  '1680x1050 (WSXGA+)'
);

UPDATE products SET max_resolution = '1080p - Full HD'
WHERE max_resolution IN (
  '1920x1080 (Full HD)',
  '2560x1080 (Ultrawide FHD)'
);

UPDATE products SET max_resolution = '1400p - 2K Quad HD'
WHERE max_resolution IN (
  '2560x1440 (QHD)',
  '3440x1440 (Ultrawide QHD)'
);

UPDATE products SET max_resolution = '2160p - 4K Ultra HD'
WHERE max_resolution = '3840x2160 (4K UHD)';
