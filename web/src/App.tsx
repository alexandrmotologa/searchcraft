import React, { useState, useEffect } from 'react';
import { Navbar, NavTab } from './components/Navbar';
import { SearchSandbox } from './components/SearchSandbox';
import { DocumentManager } from './components/DocumentManager';
import { SynonymsManager } from './components/SynonymsManager';
import { CrawlerManager } from './components/CrawlerManager';
import { QueryAnalytics } from './components/QueryAnalytics';
import { KeysManager } from './components/KeysManager';
import { WidgetEmbed } from './components/WidgetEmbed';
import { fetchHealth, fetchIndexes, createIndex, upsertDocuments } from './api/client';
import { IndexMeta, SystemHealth } from './types/search';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<NavTab>('sandbox');
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [indexes, setIndexes] = useState<IndexMeta[]>([]);
  const [selectedIndex, setSelectedIndex] = useState<string>('products');

  const loadInitialData = async () => {
    try {
      const h = await fetchHealth();
      setHealth(h);

      let idxList = await fetchIndexes();
      if (idxList.length === 0) {
        await createIndex('products', { title: 3.0, description: 1.0, tags: 2.0 }, 'hybrid');
        const sampleProducts = [
          {
            id: 'prod_1',
            title: 'Keychron Q1 Pro Wireless Custom Mechanical Keyboard',
            description: 'Compact 75% layout, CNC aluminum body, hot-swappable switches, wireless Bluetooth 5.1 and Type-C connectivity.',
            category: 'Keyboards',
            brand: 'Keychron',
            price: 199.0,
            rating: 4.9,
            in_stock: true,
            tags: ['wireless', 'aluminum', 'mechanical', 'custom'],
          },
          {
            id: 'prod_2',
            title: 'Logitech MX Master 3S Wireless Performance Mouse',
            description: 'Quiet clicks, 8K DPI any-surface tracking, MagSpeed electromagnetic scrolling wheel with ergonomic thumb rest.',
            category: 'Mice',
            brand: 'Logitech',
            price: 99.99,
            rating: 4.8,
            in_stock: true,
            tags: ['wireless', 'ergonomic', 'productivity', 'office'],
          },
          {
            id: 'prod_3',
            title: 'GMMK Pro 75% Gasket-Mounted Mechanical Keyboard',
            description: 'Rotary encoder knob, fully modular switch sockets, vibrant RGB backlighting, black anodized case.',
            category: 'Keyboards',
            brand: 'Glorious',
            price: 169.99,
            rating: 4.7,
            in_stock: true,
            tags: ['mechanical', 'rgb', 'custom', 'gasket'],
          },
          {
            id: 'prod_4',
            title: 'Razer DeathAdder V3 Pro Ultra-Lightweight Wireless Mouse',
            description: '63g ultra-lightweight ergonomic esports mouse, Focus Pro 30K optical sensor, optical switches Gen-3.',
            category: 'Mice',
            brand: 'Razer',
            price: 149.99,
            rating: 4.6,
            in_stock: true,
            tags: ['gaming', 'wireless', 'ultralight', 'fast'],
          },
          {
            id: 'prod_5',
            title: 'Ducky One 3 TKL RGB Mechanical Keyboard',
            description: 'Dual layer hotswap PCB with Kailh yellow sockets, thick PBT double-shot seamless keycaps, tuned stabilizers.',
            category: 'Keyboards',
            brand: 'Ducky',
            price: 119.0,
            rating: 4.7,
            in_stock: false,
            tags: ['mechanical', 'rgb', 'tkl', 'pbt'],
          },
        ];
        await upsertDocuments('products', sampleProducts);
        idxList = await fetchIndexes();
      }

      setIndexes(idxList);
      if (idxList.length > 0) {
        setSelectedIndex(idxList[0].name);
      }
    } catch (err) {
      console.error('Initialization error:', err);
    }
  };

  useEffect(() => {
    loadInitialData();
  }, []);

  const refreshIndexes = async () => {
    try {
      const list = await fetchIndexes();
      setIndexes(list);
      if (list.length > 0 && !list.find((i) => i.name === selectedIndex)) {
        setSelectedIndex(list[0].name);
      }
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="min-h-screen bg-background text-slate-100 flex flex-col font-sans">
      <Navbar activeTab={activeTab} setActiveTab={setActiveTab} health={health} />

      <main className="flex-1">
        {activeTab === 'sandbox' && (
          <SearchSandbox
            indexes={indexes}
            selectedIndex={selectedIndex}
            onSelectIndex={setSelectedIndex}
          />
        )}

        {activeTab === 'documents' && (
          <DocumentManager
            indexes={indexes}
            selectedIndex={selectedIndex}
            onSelectIndex={setSelectedIndex}
            onRefreshIndexes={refreshIndexes}
          />
        )}

        {activeTab === 'synonyms' && (
          <SynonymsManager
            indexes={indexes}
            selectedIndex={selectedIndex}
            onSelectIndex={setSelectedIndex}
          />
        )}

        {activeTab === 'crawler' && (
          <CrawlerManager
            indexes={indexes}
            selectedIndex={selectedIndex}
            onSelectIndex={setSelectedIndex}
            onRefreshIndexes={refreshIndexes}
          />
        )}

        {activeTab === 'analytics' && <QueryAnalytics />}

        {activeTab === 'keys' && <KeysManager />}

        {activeTab === 'widget' && (
          <WidgetEmbed indexes={indexes} selectedIndex={selectedIndex} />
        )}
      </main>

      <footer className="border-t border-border/60 py-6 text-center text-xs text-slate-500">
        <p>SearchCraft &bull; Single-Binary Instant Hybrid Search Engine (&lt; 40MB RAM)</p>
      </footer>
    </div>
  );
};
