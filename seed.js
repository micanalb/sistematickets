// seed.js — Carga datos de prueba en la base de datos sistematickets
// Solo inserta los eventos si no existen (verifica por título)
// Uso: node seed.js

const mysql = require('mysql2/promise');
const fs    = require('fs');
const path  = require('path');

// ── Leer el .env del backend ──────────────────────────────────────
function leerEnv() {
  const envPath = path.join(__dirname, 'backend', '.env');
  if (!fs.existsSync(envPath)) {
    console.error('❌ No se encontró backend/.env');
    process.exit(1);
  }
  const env = {};
  for (const linea of fs.readFileSync(envPath, 'utf-8').split('\n')) {
    const limpia = linea.trim();
    if (!limpia || limpia.startsWith('#')) continue;
    const [clave, ...resto] = limpia.split('=');
    env[clave.trim()] = resto.join('=').trim();
  }
  return env;
}

// ── Eventos con fechas FUTURAS ────────────────────────────────────
const EVENTOS = [
  {
    titulo: 'Lollapalooza Argentina 2026',
    descripcion: 'El festival de música más grande del país con artistas internacionales.',
    fecha_hora: '2026-11-15 14:00:00',
    duracion_minutos: 480,
    lugar: 'Hipódromo de Palermo',
    direccion: 'Av. del Libertador 4101',
    ciudad: 'Buenos Aires',
    categoria: 'musica',
    capacidad_total: 50000,
    entradas_vendidas: 32000,
    precio_base: 45000.00,
    imagen_url: '',
    estado: 'activo',
  },
  {
    titulo: 'River vs Boca - Superclásico',
    descripcion: 'El partido más esperado del año en el Monumental.',
    fecha_hora: '2026-12-01 21:00:00',
    duracion_minutos: 120,
    lugar: 'Estadio Monumental',
    direccion: 'Av. Figueroa Alcorta 7597',
    ciudad: 'Buenos Aires',
    categoria: 'deporte',
    capacidad_total: 84000,
    entradas_vendidas: 70000,
    precio_base: 25000.00,
    imagen_url: '',
    estado: 'activo',
  },
  {
    titulo: 'TEDx Buenos Aires 2026',
    descripcion: 'Charlas inspiradoras de líderes y pensadores de todo el mundo.',
    fecha_hora: '2026-11-20 09:00:00',
    duracion_minutos: 480,
    lugar: 'Teatro Gran Rex',
    direccion: 'Av. Corrientes 857',
    ciudad: 'Buenos Aires',
    categoria: 'conferencia',
    capacidad_total: 1200,
    entradas_vendidas: 400,
    precio_base: 12000.00,
    imagen_url: '',
    estado: 'activo',
  },
  {
    titulo: 'Carmen - Teatro Colón',
    descripcion: 'La ópera más famosa de Bizet en el teatro más importante de Latinoamérica.',
    fecha_hora: '2026-11-25 20:00:00',
    duracion_minutos: 150,
    lugar: 'Teatro Colón',
    direccion: 'Cerrito 628',
    ciudad: 'Buenos Aires',
    categoria: 'teatro_cine',
    capacidad_total: 2500,
    entradas_vendidas: 1800,
    precio_base: 35000.00,
    imagen_url: '',
    estado: 'activo',
  },
  {
    titulo: 'Festival de Jazz Córdoba',
    descripcion: 'Tres días de jazz en vivo con músicos locales e internacionales.',
    fecha_hora: '2026-12-05 18:00:00',
    duracion_minutos: 360,
    lugar: 'Plaza San Martín',
    direccion: 'Plaza San Martín s/n',
    ciudad: 'Córdoba',
    categoria: 'musica',
    capacidad_total: 5000,
    entradas_vendidas: 1200,
    precio_base: 8000.00,
    imagen_url: '',
    estado: 'activo',
  },
  {
    titulo: 'Maratón Ciudad de Buenos Aires',
    descripcion: 'La maratón anual más popular de Argentina con 42km por la ciudad.',
    fecha_hora: '2026-10-10 07:00:00',
    duracion_minutos: 360,
    lugar: 'Obelisco',
    direccion: 'Av. 9 de Julio s/n',
    ciudad: 'Buenos Aires',
    categoria: 'deporte',
    capacidad_total: 20000,
    entradas_vendidas: 18500,
    precio_base: 5000.00,
    imagen_url: '',
    estado: 'activo',
  },
];

// ── Main ──────────────────────────────────────────────────────────
async function main() {
  const env = leerEnv();

  const conexion = await mysql.createConnection({
    host:     env.DB_HOST     || 'localhost',
    port:     parseInt(env.DB_PORT || '3306'),
    user:     env.DB_USER     || 'root',
    password: env.DB_PASSWORD || '',
    database: env.DB_NAME     || 'sistematickets',
  });

  console.log(`✅ Conectado a MySQL — base: ${env.DB_NAME}`);
  console.log('🔍 Verificando eventos existentes...\n');

  let insertados = 0;
  let omitidos   = 0;

  for (const evento of EVENTOS) {
    const [filas] = await conexion.execute(
      'SELECT id FROM eventos WHERE titulo = ? AND deleted_at IS NULL LIMIT 1',
      [evento.titulo]
    );

    if (filas.length > 0) {
      console.log(`⏭️  Ya existe: "${evento.titulo}" — omitido`);
      omitidos++;
      continue;
    }

    await conexion.execute(
      `INSERT INTO eventos
        (titulo, descripcion, fecha_hora, duracion_minutos, lugar, direccion, ciudad,
         categoria, capacidad_total, entradas_vendidas, precio_base, imagen_url, estado,
         created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
      [
        evento.titulo, evento.descripcion, evento.fecha_hora, evento.duracion_minutos,
        evento.lugar, evento.direccion, evento.ciudad, evento.categoria,
        evento.capacidad_total, evento.entradas_vendidas, evento.precio_base,
        evento.imagen_url, evento.estado,
      ]
    );

    console.log(`✅ Insertado: "${evento.titulo}"`);
    insertados++;
  }

  await conexion.end();

  console.log(`\n📊 Resultado: ${insertados} insertados, ${omitidos} omitidos`);
  if (insertados === 0) {
    console.log('ℹ️  Todos los eventos ya estaban cargados. No se modificó nada.');
  } else {
    console.log('🎉 Seed completado correctamente.');
  }
}

main().catch(err => {
  console.error('❌ Error en el seed:', err.message);
  process.exit(1);
});
