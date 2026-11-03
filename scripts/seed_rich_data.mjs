const baseUrl = 'http://127.0.0.1:7700';

async function post(path, body) {
  const res = await fetch(`${baseUrl}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  return res.json();
}

async function main() {
  console.log('Seeding rich demonstration data into SearchCraft...');

  // 1. More e-commerce products
  const products = [
    {
      id: 'prod_101',
      title: 'Keychron Q1 Pro Wireless Custom Mechanical Keyboard',
      description: '75% layout QMK/VIA wireless custom mechanical keyboard with CNC aluminum body and hot-swappable switches.',
      category: 'Keyboards',
      brand: 'Keychron',
      price: 199.99,
      tags: ['wireless', 'mechanical', 'custom', 'qmk'],
      in_stock: true,
    },
    {
      id: 'prod_102',
      title: 'Logitech MX Master 3S Advanced Wireless Mouse',
      description: 'Ergonomic performance mouse with quiet clicks, 8K DPI sensor on glass, and MagSpeed electromagnetic scrolling.',
      category: 'Mice',
      brand: 'Logitech',
      price: 99.99,
      tags: ['ergonomic', 'wireless', 'office', 'bluetooth'],
      in_stock: true,
    },
    {
      id: 'prod_103',
      title: 'Sony WH-1000XM5 Wireless Noise Cancelling Headphones',
      description: 'Industry-leading noise cancellation with two processors and 8 microphones for superior audio and clear calls.',
      category: 'Audio',
      brand: 'Sony',
      price: 349.99,
      tags: ['anc', 'wireless', 'bluetooth', 'audiophile'],
      in_stock: true,
    },
    {
      id: 'prod_104',
      title: 'Dell UltraSharp U2723QE 27" 4K UHD IPS USB-C Hub Monitor',
      description: 'Brilliant 4K IPS Black panel with 2000:1 contrast ratio, 98% DCI-P3 color gamut, and 90W USB-C power delivery.',
      category: 'Monitors',
      brand: 'Dell',
      price: 579.99,
      tags: ['4k', 'ips', 'usbc', 'designer'],
      in_stock: true,
    },
    {
      id: 'prod_105',
      title: 'Apple MacBook Pro 14" M3 Pro 18GB 512GB Space Black',
      description: 'Blazing fast Apple Silicon workstation laptop with Liquid Retina XDR display and 18 hours of battery life.',
      category: 'Laptops',
      brand: 'Apple',
      price: 1999.00,
      tags: ['laptop', 'm3', 'macos', 'pro'],
      in_stock: true,
    },
    {
      id: 'prod_106',
      title: 'ZSA Moonlander Mark I Split Ergonomic Mechanical Keyboard',
      description: 'Next-generation ergonomic split keyboard with thumb clusters, hot-swap sockets, and live firmware flashing.',
      category: 'Keyboards',
      brand: 'ZSA',
      price: 365.00,
      tags: ['ergonomic', 'split', 'mechanical', 'ortholinear'],
      in_stock: true,
    },
    {
      id: 'prod_107',
      title: 'Razer DeathAdder V3 Pro Ultra-lightweight Wireless Esports Mouse',
      description: '63g ultra-lightweight ergonomic esports gaming mouse with Focus Pro 30K optical sensor and optical mouse switches.',
      category: 'Mice',
      brand: 'Razer',
      price: 149.99,
      tags: ['gaming', 'wireless', 'ultralight', '30k'],
      in_stock: true,
    },
    {
      id: 'prod_108',
      title: 'Bose QuietComfort Ultra Wireless Spatial Audio Headphones',
      description: 'World-class noise cancellation, breakthrough spatialized audio, and elevated luxury design.',
      category: 'Audio',
      brand: 'Bose',
      price: 429.00,
      tags: ['anc', 'spatial', 'wireless', 'travel'],
      in_stock: true,
    },
  ];

  await post('/api/v1/indexes/products/documents', products);
  console.log('✓ Ingested e-commerce products');

  // 2. Create docs index and populate
  await post('/api/v1/indexes', { name: 'documentation' });
  const docs = [
    {
      id: 'doc_1',
      title: 'Getting Started with Hybrid Search and RRF',
      content: 'Learn how SearchCraft blends BM25 term scores and dense vector embeddings using Reciprocal Rank Fusion parameter k=60.',
      category: 'Guides',
      tags: ['hybrid', 'rrf', 'bm25', 'vectors'],
    },
    {
      id: 'doc_2',
      title: 'Configuring Typo Tolerance with Damerau-Levenshtein',
      content: 'Configure adaptive typo tolerance rules, character transpositions, and fuzzy query candidate expansions.',
      category: 'Reference',
      tags: ['typo', 'levenshtein', 'algorithms'],
    },
    {
      id: 'doc_3',
      title: 'Embedding Providers: Fast Local, ONNX, and OpenAI',
      content: 'Switch between the built-in 256-dimension n-gram vectorizer, local ONNX Runtime all-MiniLM-L6-v2, and OpenAI embeddings.',
      category: 'Architecture',
      tags: ['onnx', 'embeddings', 'models'],
    },
    {
      id: 'doc_4',
      title: 'Embedding the 2KB Cmd+K Palette into Web Apps',
      content: 'Include the single script tag on any static or dynamic website to provide an instant modal command palette.',
      category: 'Integration',
      tags: ['widget', 'cmdk', 'palette', 'vanilla'],
    },
  ];
  await post('/api/v1/indexes/documentation/documents', docs);
  console.log('✓ Ingested documentation records');

  // 3. Add Synonyms
  await post('/api/v1/indexes/products/synonyms', {
    term: 'notebook',
    synonyms: ['laptop', 'macbook'],
    one_way: false,
  });
  await post('/api/v1/indexes/products/synonyms', {
    term: 'display',
    synonyms: ['monitor', 'screen'],
    one_way: false,
  });
  console.log('✓ Added synonyms');

  // 4. Add Merchandising Promotion
  await post('/api/v1/indexes/products/promotions', {
    query: 'keyboard',
    document_id: 'prod_106',
    position: 0,
    label: 'Featured Ergonomic Pick',
  });
  console.log('✓ Added promotions');

  // 5. Run realistic searches to generate analytics
  const queries = [
    'ergonomic keyboard',
    'wireless mouse',
    'noise cancelling',
    '4k monitor',
    'macbook pro',
    'mechanical switches',
    'split ergonomic',
    'wireless earbuds', // might return 0
    'ergonomic',
    'laptop',
  ];

  for (const q of queries) {
    await post('/api/v1/indexes/products/search', {
      query: q,
      mode: 'hybrid',
      limit: 10,
      hybrid_alpha: 0.5,
      highlight: true,
      facets: ['category', 'brand', 'tags'],
    });
  }
  console.log('✓ Generated query logs for analytics');
}

main().catch(console.error);
