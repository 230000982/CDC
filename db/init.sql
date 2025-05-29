CREATE TABLE cargo (
    id_cargo INT PRIMARY KEY AUTO_INCREMENT,
    descricao TEXT NOT NULL
);

CREATE TABLE user (
    id_user INT PRIMARY KEY AUTO_INCREMENT,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    cargo_id INT,
    failed_attempts INT DEFAULT 0,
    FOREIGN KEY (cargo_id) REFERENCES cargo(id_cargo)
);

CREATE TABLE tipo (
    id_tipo INT PRIMARY KEY AUTO_INCREMENT,
    descricao VARCHAR(255) NOT NULL
);

CREATE TABLE plataforma (
    id_platforma INT PRIMARY KEY AUTO_INCREMENT,
    descricao VARCHAR(255) NOT NULL
);

CREATE TABLE estado (
    id_estado INT PRIMARY KEY AUTO_INCREMENT,
    descricao VARCHAR(255) NOT NULL
);

CREATE TABLE resultado (
    id_resultado INT PRIMARY KEY AUTO_INCREMENT,
    descricao VARCHAR(255) NOT NULL
);

CREATE TABLE logs (
    id_logs INT PRIMARY KEY AUTO_INCREMENT,
    tabela VARCHAR(255) NOT NULL,
    acao VARCHAR(255) NOT NULL,
    old_data TEXT,
    new_data TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    id_user INT,
    FOREIGN KEY (id_user) REFERENCES user(id_user)
);

CREATE TABLE concurso (
    id_concurso INT PRIMARY KEY AUTO_INCREMENT,
    referencia VARCHAR(255),
    entidade VARCHAR(255),
    dia_erro VARCHAR(255),
    hora_erro VARCHAR(255),
    dia_proposta VARCHAR(255),
    hora_proposta VARCHAR(255),
    preco DECIMAL(11, 2),
    tipo_id INT,
    plataforma_id INT,
    referencia_bc VARCHAR(255),
    preliminar BOOLEAN,   
    dia_audiencia VARCHAR(255),
    hora_audiencia VARCHAR(255),
    final BOOLEAN,
    recurso BOOLEAN,
    impugnacao BOOLEAN,
    estado_id INT,
    link VARCHAR(255),
    adjudicatario VARCHAR(255),
    resultado_id INT,
    FOREIGN KEY (resultado_id) REFERENCES resultado(id_resultado),
    FOREIGN KEY (tipo_id) REFERENCES tipo(id_tipo),
    FOREIGN KEY (plataforma_id) REFERENCES plataforma(id_platforma),
    FOREIGN KEY (estado_id) REFERENCES estado(id_estado)
);

INSERT INTO resultado (id_resultado, descricao) VALUES (1, '');
INSERT INTO resultado (id_resultado, descricao) VALUES (2, 'Ganho');
INSERT INTO resultado (id_resultado, descricao) VALUES (3, 'Concorrência');

INSERT INTO cargo (id_cargo, descricao) VALUES (1, 'admin');
INSERT INTO cargo (id_cargo, descricao) VALUES (2, 'SAV');
INSERT INTO cargo (id_cargo, descricao) VALUES (3, 'DCP');
INSERT INTO cargo (id_cargo, descricao) VALUES (4, 'GUEST');

INSERT INTO tipo (id_tipo, descricao) VALUES (1, '');
INSERT INTO tipo (id_tipo, descricao) VALUES (2, 'CTE');
INSERT INTO tipo (id_tipo, descricao) VALUES (3, 'CON');
INSERT INTO tipo (id_tipo, descricao) VALUES (4, 'INF');
INSERT INTO tipo (id_tipo, descricao) VALUES (5, 'CI');
INSERT INTO tipo (id_tipo, descricao) VALUES (6, 'ROB');

INSERT INTO plataforma (id_platforma, descricao) VALUES (1, '');
INSERT INTO plataforma (id_platforma, descricao) VALUES (2, 'email');
INSERT INTO plataforma (id_platforma, descricao) VALUES (3, 'vortal');
INSERT INTO plataforma (id_platforma, descricao) VALUES (4, 'acingov');
INSERT INTO plataforma (id_platforma, descricao) VALUES (5, 'anogov');
INSERT INTO plataforma (id_platforma, descricao) VALUES (6, 'saphety');

INSERT INTO user (nome, email, password, cargo_id) 
VALUES ('beso', 'notbeso2000@gmail.com', 'beso', 1);

INSERT INTO estado (id_estado, descricao) VALUES (1, '');
INSERT INTO estado (id_estado, descricao) VALUES (2, 'Em Andamento');
INSERT INTO estado (id_estado, descricao) VALUES (3, 'Enviado');
INSERT INTO estado (id_estado, descricao) VALUES (4, 'Não Enviado');
INSERT INTO estado (id_estado, descricao) VALUES (5, 'Declaração');

INSERT INTO concurso (
    referencia, entidade, dia_erro, hora_erro, dia_proposta, hora_proposta, preco, 
    tipo_id, plataforma_id, referencia_bc, preliminar, dia_audiencia, hora_audiencia, 
    final, recurso, impugnacao, estado_id, link, adjudicatario, resultado_id
) 
VALUES (
    '987654321',                         -- referencia
    'Câmara Municipal de Lisboa',       -- entidade
    '2025-04-10',                       -- dia_erro
    '09:30',                            -- hora_erro
    '2025-04-12',                       -- dia_proposta
    '15:00',                            -- hora_proposta
    25000.00,                           -- preco
    2,                                  -- tipo_id (CTE)
    3,                                  -- plataforma_id (vortal)
    'BC5678',                           -- referencia_bc
    TRUE,                               -- preliminar
    '2025-04-14',                       -- dia_audiencia
    '10:00',                            -- hora_audiencia
    FALSE,                              -- final
    FALSE,                              -- recurso
    FALSE,                              -- impugnacao
    2,                                  -- estado_id (Em Andamento)
    'https://concursos.gov.pt/987654321', -- link
    'notbeso2000@gmail.com',            -- adjudicatario
    2                                   -- resultado_id (Ganho)
);
