-- Seed data for servents
INSERT IGNORE INTO servents (name, hostname, port, auth_id, passwd, priority, max_channels, enabled, yellow_pages) VALUES
('Localテスト用Peercast', '192.168.50.11', 7154, 'admin', 'hogehoge', 0, 3, 1, '0yp');

-- Seed data for notices
INSERT INTO notices (title, body) VALUES
('メンテナンスのお知らせ', '定期メンテナンスを実施します。');
