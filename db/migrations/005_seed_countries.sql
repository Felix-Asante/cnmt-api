-- +goose Up
INSERT INTO countries (
        name,
        iso_code,
        flag,
        currency_name,
        currency_code,
        currency_symbol
    )
VALUES (
        'Ghana',
        'GH',
        '🇬🇭',
        'Ghanaian Cedi',
        'GHS',
        '₵'
    ),
    (
        'Liberia',
        'LR',
        '🇱🇷',
        'Liberian Dollar',
        'LRD',
        '$'
    ),
    (
        'Sierra Leone',
        'SL',
        '🇸🇱',
        'Leone',
        'SLE',
        'Le'
    ),
    (
        'Maroc',
        'MA',
        '🇲🇦',
        'Moroccan Dirham',
        'MAD',
        'DH'
    ),
    (
        'Mozambique',
        'MZ',
        '🇲🇿',
        'Mozambican Metical',
        'MZN',
        'MT'
    ),
    (
        'Kenya',
        'KE',
        '🇰🇪',
        'Kenyan Shilling',
        'KES',
        'KSh'
    ),
    (
        'Côte d''Ivoire',
        'CI',
        '🇨🇮',
        'West African CFA Franc',
        'XOF',
        'CFA'
    ),
    (
        'Mali',
        'ML',
        '🇲🇱',
        'West African CFA Franc',
        'XOF',
        'CFA'
    ),
    (
        'Burkina Faso',
        'BF',
        '🇧🇫',
        'West African CFA Franc',
        'XOF',
        'CFA'
    ),
    (
        'Togo',
        'TG',
        '🇹🇬',
        'West African CFA Franc',
        'XOF',
        'CFA'
    ),
    (
        'Benin',
        'BJ',
        '🇧🇯',
        'West African CFA Franc',
        'XOF',
        'CFA'
    ),
    (
        'Cameroun',
        'CM',
        '🇨🇲',
        'Central African CFA Franc',
        'XAF',
        'CFA'
    ),
    (
        'Congo',
        'CG',
        '🇨🇬',
        'Central African CFA Franc',
        'XAF',
        'CFA'
    ),
    (
        'Europe',
        'EU',
        '🇪🇺',
        'Euro',
        'EUR',
        '€'
    );
INSERT INTO payment_channels (name, channel_type, country_id)
SELECT v.name,
    v.channel_type::receiving_methods,
    c.id
FROM (
        VALUES -- Ghana
            ('GH', 'MTN', 'MOBILE_MONEY'),
            ('GH', 'Telecel', 'MOBILE_MONEY'),
            ('GH', 'Ecobank', 'BANK'),
            ('GH', 'Absa', 'BANK'),
            ('GH', 'GCB', 'BANK'),
            ('GH', 'Bank transfer', 'BANK'),
            -- Liberia
            ('LR', 'MTN', 'MOBILE_MONEY'),
            ('LR', 'Orange', 'MOBILE_MONEY'),
            -- Sierra Leone
            ('SL', 'Orange', 'MOBILE_MONEY'),
            -- Maroc
            ('MA', 'CIH', 'BANK'),
            ('MA', 'Attijariwafa', 'BANK'),
            ('MA', 'Banque Populaire', 'BANK'),
            -- Mozambique
            ('MZ', 'm-pesa (Vodafone)', 'MOBILE_MONEY'),
            -- Kenya
            ('KE', 'm-pesa Kenya', 'MOBILE_MONEY'),
            -- Côte d'Ivoire, Mali, Burkina Faso, Togo, Benin
            ('CI', 'MTN', 'MOBILE_MONEY'),
            ('CI', 'Orange', 'MOBILE_MONEY'),
            ('CI', 'Moov', 'MOBILE_MONEY'),
            ('CI', 'Wave', 'MOBILE_MONEY'),
            ('ML', 'MTN', 'MOBILE_MONEY'),
            ('ML', 'Orange', 'MOBILE_MONEY'),
            ('ML', 'Moov', 'MOBILE_MONEY'),
            ('ML', 'Wave', 'MOBILE_MONEY'),
            ('BF', 'MTN', 'MOBILE_MONEY'),
            ('BF', 'Orange', 'MOBILE_MONEY'),
            ('BF', 'Moov', 'MOBILE_MONEY'),
            ('BF', 'Wave', 'MOBILE_MONEY'),
            ('TG', 'MTN', 'MOBILE_MONEY'),
            ('TG', 'Orange', 'MOBILE_MONEY'),
            ('TG', 'Moov', 'MOBILE_MONEY'),
            ('TG', 'Wave', 'MOBILE_MONEY'),
            ('BJ', 'MTN', 'MOBILE_MONEY'),
            ('BJ', 'Orange', 'MOBILE_MONEY'),
            ('BJ', 'Moov', 'MOBILE_MONEY'),
            ('BJ', 'Wave', 'MOBILE_MONEY'),
            -- Cameroun, Congo
            ('CM', 'MTN', 'MOBILE_MONEY'),
            ('CM', 'Orange', 'MOBILE_MONEY'),
            ('CG', 'MTN', 'MOBILE_MONEY'),
            ('CG', 'Orange', 'MOBILE_MONEY'),
            -- Europe
            ('EU', 'IBAN', 'BANK')
    ) AS v(iso_code, name, channel_type)
    JOIN countries c ON c.iso_code = v.iso_code
    AND c.deleted_at IS NULL;
-- +goose Down
DELETE FROM payment_channels
WHERE country_id IN (
        SELECT id
        FROM countries
        WHERE iso_code IN (
                'GH',
                'LR',
                'SL',
                'MA',
                'MZ',
                'KE',
                'CI',
                'ML',
                'BF',
                'TG',
                'BJ',
                'CM',
                'CG',
                'EU'
            )
    );
DELETE FROM countries
WHERE iso_code IN (
        'GH',
        'LR',
        'SL',
        'MA',
        'MZ',
        'KE',
        'CI',
        'ML',
        'BF',
        'TG',
        'BJ',
        'CM',
        'CG',
        'EU'
    );