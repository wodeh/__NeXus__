'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { Tv, Wifi, WifiOff, ArrowLeft, Home, Film, Tv as TvIcon, Phone, Globe, UtensilsCrossed, Sparkles, MessageSquare, Settings, ChevronRight, Volume2, Moon, Sun, Play, Search } from 'lucide-react';

// --- Types ---
interface GuestInfo {
  guestName: string;
  roomNumber: string;
  checkIn: string;
  checkOut: string;
  nights: number;
  welcomeMessage: string;
  wifiName: string;
  wifiPassword: string;
  language: string;
}

interface Channel {
  id: string;
  name: string;
  number: number;
  category: string;
  logo?: string;
  isPremium: boolean;
}

interface ContentItem {
  id: string;
  title: string;
  type: 'movie' | 'series' | 'info' | 'music';
  category: string;
  description: string;
  duration?: number;
  thumbnail?: string;
}

interface MenuItem {
  id: string;
  name: string;
  price: number;
  category: string;
  description: string;
  image?: string;
}

interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'warning' | 'urgent';
  time: string;
  read: boolean;
}

type Screen = 'welcome' | 'channels' | 'content' | 'services' | 'info' | 'settings' | 'notifications' | 'youtube';

// --- Demo Data ---
const DEMO_GUEST: GuestInfo = {
  guestName: 'Alex Mercer',
  roomNumber: '1005',
  checkIn: '2026-05-12',
  checkOut: '2026-05-15',
  nights: 3,
  welcomeMessage: 'Welcome to Grand Plaza Hotel',
  wifiName: 'GrandPlaza-Guest',
  wifiPassword: 'Welcome2026',
  language: 'en',
};

const DEMO_CHANNELS: Channel[] = [
  { id: '1', name: 'Nexus Welcome', number: 1, category: 'welcome', isPremium: false },
  { id: '2', name: 'CNN International', number: 2, category: 'news', isPremium: false },
  { id: '3', name: 'BBC World', number: 3, category: 'news', isPremium: false },
  { id: '4', name: 'Al Jazeera', number: 4, category: 'news', isPremium: false },
  { id: '10', name: 'ESPN', number: 10, category: 'sports', isPremium: true },
  { id: '11', name: 'beIN Sports', number: 11, category: 'sports', isPremium: true },
  { id: '20', name: 'HBO', number: 20, category: 'movies', isPremium: true },
  { id: '21', name: 'Netflix', number: 21, category: 'movies', isPremium: true },
  { id: '30', name: 'Spa TV', number: 30, category: 'wellness', isPremium: false },
  { id: '50', name: 'Cartoon Network', number: 50, category: 'kids', isPremium: false },
  { id: '80', name: 'Music FM', number: 80, category: 'music', isPremium: false },
];

const DEMO_CONTENT: ContentItem[] = [
  { id: '1', title: 'Local Attractions', type: 'info', category: 'tourism', description: 'Top places to visit within 5km' },
  { id: '2', title: 'Room Service Menu', type: 'info', category: 'dining', description: 'Full dining menu with ordering' },
  { id: '3', title: 'Spa & Wellness', type: 'info', category: 'wellness', description: 'Spa treatments and booking' },
  { id: '4', title: 'The Dark Knight', type: 'movie', category: 'action', description: 'Batman faces the Joker', duration: 152 },
  { id: '5', title: 'Inception', type: 'movie', category: 'sci-fi', description: 'Dream-sharing technology', duration: 148 },
];

const DEMO_MENU: MenuItem[] = [
  { id: '1', name: 'Caesar Salad', price: 18, category: 'starters', description: 'Classic with parmesan' },
  { id: '2', name: 'Grilled Salmon', price: 32, category: 'mains', description: 'With lemon butter sauce' },
  { id: '3', name: 'Ribeye Steak', price: 45, category: 'mains', description: '12oz with roasted vegetables' },
  { id: '4', name: 'Chocolate Fondant', price: 14, category: 'desserts', description: 'With vanilla ice cream' },
];

const DEMO_NOTIFICATIONS: Notification[] = [
  { id: '1', title: 'Welcome!', message: 'Enjoy your stay at Grand Plaza', type: 'info', time: '15:00', read: false },
  { id: '2', title: 'Spa Offer', message: '20% off massages today', type: 'info', time: '16:30', read: false },
];

// --- Components ---

export default function GuestTVPage() {
  const params = useParams();
  const roomId = params.roomId as string || '1005';
  
  const [currentScreen, setCurrentScreen] = useState<Screen>('welcome');
  const [guest, setGuest] = useState<GuestInfo>(DEMO_GUEST);
  const [channels] = useState<Channel[]>(DEMO_CHANNELS);
  const [content] = useState<ContentItem[]>(DEMO_CONTENT);
  const [menu] = useState<MenuItem[]>(DEMO_MENU);
  const [notifications, setNotifications] = useState<Notification[]>(DEMO_NOTIFICATIONS);
  const [selectedChannel, setSelectedChannel] = useState<Channel | null>(null);
  const [selectedContent, setSelectedContent] = useState<ContentItem | null>(null);
  const [language, setLanguage] = useState('en');
  const [largeText, setLargeText] = useState(false);
  const [highContrast, setHighContrast] = useState(false);
  const [cart, setCart] = useState<MenuItem[]>([]);
  const [currentTime, setCurrentTime] = useState(new Date());
  
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 60000);
    return () => clearInterval(timer);
  }, []);

  const unreadCount = notifications.filter(n => !n.read).length;

  const markAllRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
  };

  const addToCart = (item: MenuItem) => {
    setCart(prev => [...prev, item]);
  };

  const removeFromCart = (index: number) => {
    setCart(prev => prev.filter((_, i) => i !== index));
  };

  const cartTotal = cart.reduce((sum, item) => sum + item.price, 0);

  // TV-optimized scale factor
  const scale = largeText ? 'scale-110' : 'scale-100';
  const contrastClass = highContrast ? 'bg-black text-white' : 'bg-slate-900 text-white';

  return (
    <div className={`min-h-screen ${contrastClass} overflow-hidden select-none transition-all duration-300`}>
      {/* Top Bar */}
      <div className="flex items-center justify-between px-8 py-4 border-b border-slate-700/50">
        <div className="flex items-center gap-4">
          <TvIcon className="h-6 w-6 text-nexus-400" />
          <span className="text-lg font-bold">Nexus TV</span>
          {unreadCount > 0 && (
            <span className="flex h-5 w-5 items-center justify-center rounded-full bg-nexus-500 text-[10px] font-bold">
              {unreadCount}
            </span>
          )}
        </div>
        <div className="flex items-center gap-6 text-sm text-slate-400">
          <span>{currentTime.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
          <span>Room {guest.roomNumber}</span>
          <Wifi className="h-4 w-4 text-emerald-400" />
        </div>
      </div>

      {/* Main Content */}
      <div className={`transition-all duration-300 ${scale} origin-top`}>
        {currentScreen === 'welcome' && (
          <WelcomeScreen 
            guest={guest} 
            onNavigate={setCurrentScreen}
            unreadCount={unreadCount}
          />
        )}
        
        {currentScreen === 'channels' && (
          <ChannelsScreen 
            channels={channels}
            selected={selectedChannel}
            onSelect={setSelectedChannel}
            onBack={() => setCurrentScreen('welcome')}
          />
        )}
        
        {currentScreen === 'content' && (
          <ContentScreen 
            content={content}
            selected={selectedContent}
            onSelect={setSelectedContent}
            onBack={() => setCurrentScreen('welcome')}
          />
        )}
        
        {currentScreen === 'services' && (
          <ServicesScreen 
            menu={menu}
            cart={cart}
            onAddToCart={addToCart}
            onRemoveFromCart={removeFromCart}
            cartTotal={cartTotal}
            onBack={() => setCurrentScreen('welcome')}
          />
        )}
        
        {currentScreen === 'info' && (
          <InfoScreen guest={guest} onBack={() => setCurrentScreen('welcome')} />
        )}
        
        {currentScreen === 'notifications' && (
          <NotificationsScreen 
            notifications={notifications}
            onMarkAllRead={markAllRead}
            onBack={() => setCurrentScreen('welcome')}
          />
        )}
        
        {currentScreen === 'youtube' && (
          <YouTubeScreen onBack={() => setCurrentScreen('welcome')} />
        )}
      </div>
    </div>
  );
}

// --- Screen Components ---

function WelcomeScreen({ guest, onNavigate, unreadCount }: { 
  guest: GuestInfo; 
  onNavigate: (s: Screen) => void;
  unreadCount: number;
}) {
  const menuItems = [
    { id: 'channels' as Screen, label: 'Live TV', icon: TvIcon, color: 'bg-nexus-500/20 text-nexus-400' },
    { id: 'content' as Screen, label: 'Movies & Shows', icon: Film, color: 'bg-violet-500/20 text-violet-400' },
    { id: 'services' as Screen, label: 'Room Service', icon: UtensilsCrossed, color: 'bg-amber-500/20 text-amber-400' },
    { id: 'info' as Screen, label: 'Hotel Info', icon: Sparkles, color: 'bg-emerald-500/20 text-emerald-400' },
    { id: 'notifications' as Screen, label: 'Messages', icon: MessageSquare, color: 'bg-sky-500/20 text-sky-400', badge: unreadCount },
    { id: 'youtube' as Screen, label: 'YouTube', icon: Play, color: 'bg-rose-500/20 text-rose-400' },
    { id: 'settings' as Screen, label: 'Settings', icon: Settings, color: 'bg-slate-500/20 text-slate-400' },
  ];

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] px-8">
      {/* Hero Welcome */}
      <div className="text-center mb-12">
        <h1 className="text-5xl font-bold mb-3 bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
          {guest.welcomeMessage}
        </h1>
        <p className="text-xl text-slate-400">
          {guest.guestName ? `Welcome, ${guest.guestName}` : 'Welcome to your stay'}
        </p>
        <div className="mt-4 flex items-center justify-center gap-4 text-sm text-slate-500">
          <span>Check-in: {guest.checkIn}</span>
          <span>·</span>
          <span>Check-out: {guest.checkOut}</span>
          <span>·</span>
          <span>{guest.nights} nights</span>
        </div>
      </div>

      {/* WiFi Info Card */}
      <div className="mb-10 rounded-xl bg-slate-800/50 border border-slate-700/50 p-4 flex items-center gap-6">
        <Wifi className="h-5 w-5 text-emerald-400" />
        <div>
          <p className="text-xs text-slate-500">WiFi Access</p>
          <p className="text-sm font-medium">{guest.wifiName} · Password: {guest.wifiPassword}</p>
        </div>
      </div>

      {/* Main Menu Grid */}
      <div className="grid grid-cols-3 gap-4 max-w-3xl w-full">
        {menuItems.map((item) => (
          <button
            key={item.id}
            onClick={() => onNavigate(item.id)}
            className="group relative rounded-xl bg-slate-800/50 border border-slate-700/50 p-6 hover:bg-slate-700/50 hover:border-nexus-500/50 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-nexus-500"
          >
            <div className={`flex h-12 w-12 items-center justify-center rounded-lg ${item.color} mb-3`}>
              <item.icon className="h-6 w-6" />
            </div>
            <span className="text-sm font-medium">{item.label}</span>
            {item.badge ? (
              <span className="absolute top-2 right-2 flex h-5 w-5 items-center justify-center rounded-full bg-nexus-500 text-[10px] font-bold">
                {item.badge}
              </span>
            ) : null}
          </button>
        ))}
      </div>
    </div>
  );
}

function ChannelsScreen({ channels, selected, onSelect, onBack }: {
  channels: Channel[];
  selected: Channel | null;
  onSelect: (c: Channel) => void;
  onBack: () => void;
}) {
  const categories = ['all', ...new Set(channels.map(c => c.category))];
  const [activeCategory, setActiveCategory] = useState('all');
  
  const filtered = activeCategory === 'all' 
    ? channels 
    : channels.filter(c => c.category === activeCategory);

  return (
    <div className="p-8">
      <div className="flex items-center gap-4 mb-6">
        <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h2 className="text-2xl font-bold">Live TV</h2>
      </div>

      {/* Category Filter */}
      <div className="flex gap-2 mb-6">
        {categories.map(cat => (
          <button
            key={cat}
            onClick={() => setActiveCategory(cat)}
            className={`rounded-lg px-4 py-2 text-sm capitalize transition-colors ${
              activeCategory === cat 
                ? 'bg-nexus-500/20 text-nexus-400 border border-nexus-500/50' 
                : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      {/* Channel Grid */}
      <div className="grid grid-cols-4 gap-4">
        {filtered.map(ch => (
          <button
            key={ch.id}
            onClick={() => onSelect(ch)}
            className={`group relative rounded-xl border p-4 text-left transition-all hover:scale-105 focus:outline-none focus:ring-2 focus:ring-nexus-500 ${
              selected?.id === ch.id 
                ? 'border-nexus-500 bg-nexus-500/10' 
                : 'border-slate-700 bg-slate-800/50 hover:border-slate-600'
            }`}
          >
            <div className="flex items-start justify-between mb-3">
              <span className="flex h-8 w-8 items-center justify-center rounded bg-slate-700 text-sm font-bold">
                {ch.number}
              </span>
              {ch.isPremium && (
                <span className="rounded-full bg-amber-500/20 px-2 py-0.5 text-[10px] text-amber-400">
                  Premium
                </span>
              )}
            </div>
            <p className="font-medium">{ch.name}</p>
            <p className="text-xs text-slate-500 capitalize">{ch.category}</p>
          </button>
        ))}
      </div>

      {/* Now Playing Bar */}
      {selected && (
        <div className="fixed bottom-0 left-0 right-0 bg-slate-900/95 border-t border-slate-700 p-4">
          <div className="flex items-center justify-between max-w-4xl mx-auto">
            <div className="flex items-center gap-4">
              <div className="flex h-10 w-10 items-center justify-center rounded bg-nexus-500/20 text-nexus-400 font-bold">
                {selected.number}
              </div>
              <div>
                <p className="font-medium">{selected.name}</p>
                <p className="text-xs text-slate-500">Now Playing</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <Volume2 className="h-5 w-5 text-slate-400" />
              <button className="rounded-lg bg-nexus-500 px-4 py-2 text-sm font-medium hover:bg-nexus-600">
                Watch
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function ContentScreen({ content, selected, onSelect, onBack }: {
  content: ContentItem[];
  selected: ContentItem | null;
  onSelect: (c: ContentItem) => void;
  onBack: () => void;
}) {
  const categories = ['all', ...new Set(content.map(c => c.category))];
  const [activeCategory, setActiveCategory] = useState('all');
  
  const filtered = activeCategory === 'all'
    ? content
    : content.filter(c => c.category === activeCategory);

  return (
    <div className="p-8">
      <div className="flex items-center gap-4 mb-6">
        <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h2 className="text-2xl font-bold">Movies & Shows</h2>
      </div>

      <div className="flex gap-2 mb-6">
        {categories.map(cat => (
          <button
            key={cat}
            onClick={() => setActiveCategory(cat)}
            className={`rounded-lg px-4 py-2 text-sm capitalize transition-colors ${
              activeCategory === cat 
                ? 'bg-violet-500/20 text-violet-400 border border-violet-500/50' 
                : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-3 gap-4">
        {filtered.map(item => (
          <button
            key={item.id}
            onClick={() => onSelect(item)}
            className={`group rounded-xl border overflow-hidden text-left transition-all hover:scale-105 focus:outline-none focus:ring-2 focus:ring-violet-500 ${
              selected?.id === item.id 
                ? 'border-violet-500' 
                : 'border-slate-700 hover:border-slate-600'
            }`}
          >
            <div className="h-32 bg-slate-700/50 flex items-center justify-center">
              {item.type === 'movie' ? <Film className="h-10 w-10 text-violet-400" /> : 
               item.type === 'series' ? <TvIcon className="h-10 w-10 text-violet-400" /> :
               <Sparkles className="h-10 w-10 text-sky-400" />}
            </div>
            <div className="p-4">
              <p className="font-medium">{item.title}</p>
              <p className="text-xs text-slate-500">{item.category} {item.duration && `· ${item.duration}min`}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

function ServicesScreen({ menu, cart, onAddToCart, onRemoveFromCart, cartTotal, onBack }: {
  menu: MenuItem[];
  cart: MenuItem[];
  onAddToCart: (item: MenuItem) => void;
  onRemoveFromCart: (index: number) => void;
  cartTotal: number;
  onBack: () => void;
}) {
  const [showCart, setShowCart] = useState(false);
  const categories = ['all', ...new Set(menu.map(m => m.category))];
  const [activeCategory, setActiveCategory] = useState('all');
  
  const filtered = activeCategory === 'all'
    ? menu
    : menu.filter(m => m.category === activeCategory);

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
            <ArrowLeft className="h-5 w-5" />
          </button>
          <h2 className="text-2xl font-bold">Room Service</h2>
        </div>
        <button 
          onClick={() => setShowCart(!showCart)}
          className="relative rounded-lg bg-amber-500/20 px-4 py-2 text-sm font-medium text-amber-400"
        >
          Cart ({cart.length}) · ${cartTotal}
        </button>
      </div>

      {showCart ? (
        <div className="rounded-xl border border-slate-700 bg-slate-800/50 p-6">
          <h3 className="text-lg font-bold mb-4">Your Order</h3>
          {cart.length === 0 ? (
            <p className="text-slate-500">Your cart is empty</p>
          ) : (
            <div className="space-y-3">
              {cart.map((item, i) => (
                <div key={i} className="flex items-center justify-between">
                  <div>
                    <p className="font-medium">{item.name}</p>
                    <p className="text-xs text-slate-500">{item.description}</p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="font-medium">${item.price}</span>
                    <button 
                      onClick={() => onRemoveFromCart(i)}
                      className="rounded p-1 text-rose-400 hover:bg-rose-500/20"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              ))}
              <div className="border-t border-slate-700 pt-3 flex justify-between">
                <span className="font-bold">Total</span>
                <span className="font-bold">${cartTotal}</span>
              </div>
              <button className="w-full rounded-lg bg-amber-500 py-3 font-medium text-slate-900 hover:bg-amber-400">
                Place Order · Bill to Room
              </button>
            </div>
          )}
        </div>
      ) : (
        <>
          <div className="flex gap-2 mb-6">
            {categories.map(cat => (
              <button
                key={cat}
                onClick={() => setActiveCategory(cat)}
                className={`rounded-lg px-4 py-2 text-sm capitalize transition-colors ${
                  activeCategory === cat 
                    ? 'bg-amber-500/20 text-amber-400 border border-amber-500/50' 
                    : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
                }`}
              >
                {cat}
              </button>
            ))}
          </div>

          <div className="space-y-3">
            {filtered.map(item => (
              <div key={item.id} className="flex items-center justify-between rounded-xl border border-slate-700 bg-slate-800/50 p-4">
                <div>
                  <p className="font-medium">{item.name}</p>
                  <p className="text-sm text-slate-500">{item.description}</p>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-lg font-bold text-amber-400">${item.price}</span>
                  <button 
                    onClick={() => onAddToCart(item)}
                    className="rounded-lg bg-amber-500/20 px-4 py-2 text-sm font-medium text-amber-400 hover:bg-amber-500/30"
                  >
                    Add
                  </button>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

function InfoScreen({ guest, onBack }: { guest: GuestInfo; onBack: () => void }) {
  const sections = [
    { title: 'Hotel Amenities', items: ['Pool: 6AM - 10PM', 'Gym: 24/7', 'Spa: 9AM - 9PM', 'Business Center: 24/7'] },
    { title: 'Dining', items: ['The Grand Restaurant: 6AM - 11PM', 'Rooftop Bar: 4PM - 2AM', 'Poolside Café: 10AM - 6PM'] },
    { title: 'Services', items: ['Concierge: Dial 0', 'Housekeeping: Dial 1', 'Room Service: Dial 2', 'Emergency: Dial 911'] },
    { title: 'Local Info', items: ['Airport: 25km · 30min', 'City Center: 5km · 10min', 'Nearest Hospital: 3km'] },
  ];

  return (
    <div className="p-8">
      <div className="flex items-center gap-4 mb-6">
        <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h2 className="text-2xl font-bold">Hotel Information</h2>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {sections.map(section => (
          <div key={section.title} className="rounded-xl border border-slate-700 bg-slate-800/50 p-5">
            <h3 className="text-lg font-bold mb-3 text-emerald-400">{section.title}</h3>
            <ul className="space-y-2">
              {section.items.map((item, i) => (
                <li key={i} className="text-sm text-slate-400">{item}</li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="mt-6 rounded-xl border border-slate-700 bg-slate-800/50 p-5">
        <h3 className="text-lg font-bold mb-3 text-emerald-400">WiFi Access</h3>
        <div className="flex items-center gap-8">
          <div>
            <p className="text-xs text-slate-500">Network</p>
            <p className="text-xl font-medium">{guest.wifiName}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Password</p>
            <p className="text-xl font-medium">{guest.wifiPassword}</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function NotificationsScreen({ notifications, onMarkAllRead, onBack }: {
  notifications: Notification[];
  onMarkAllRead: () => void;
  onBack: () => void;
}) {
  return (
    <div className="p-8 max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
            <ArrowLeft className="h-5 w-5" />
          </button>
          <h2 className="text-2xl font-bold">Messages</h2>
        </div>
        <button 
          onClick={onMarkAllRead}
          className="text-sm text-sky-400 hover:text-sky-300"
        >
          Mark all read
        </button>
      </div>

      <div className="space-y-3">
        {notifications.map(n => (
          <div 
            key={n.id} 
            className={`rounded-xl border p-4 ${
              n.read 
                ? 'border-slate-700 bg-slate-800/30' 
                : 'border-sky-500/50 bg-sky-500/10'
            }`}
          >
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <span className={`h-2 w-2 rounded-full ${
                    n.type === 'urgent' ? 'bg-rose-500' : 
                    n.type === 'warning' ? 'bg-amber-500' : 'bg-sky-500'
                  }`} />
                  <p className="font-medium">{n.title}</p>
                </div>
                <p className="text-sm text-slate-500 mt-1">{n.message}</p>
              </div>
              <span className="text-xs text-slate-600">{n.time}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function SettingsScreen({ 
  language, onLanguageChange, 
  largeText, onLargeTextChange,
  highContrast, onHighContrastChange,
  onBack 
}: {
  language: string;
  onLanguageChange: (l: string) => void;
  largeText: boolean;
  onLargeTextChange: (v: boolean) => void;
  highContrast: boolean;
  onHighContrastChange: (v: boolean) => void;
  onBack: () => void;
}) {
  const languages = [
    { code: 'en', name: 'English' },
    { code: 'ar', name: 'العربية' },
    { code: 'fr', name: 'Français' },
    { code: 'es', name: 'Español' },
    { code: 'de', name: 'Deutsch' },
    { code: 'zh', name: '中文' },
  ];

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <div className="flex items-center gap-4 mb-6">
        <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h2 className="text-2xl font-bold">Settings</h2>
      </div>

      <div className="space-y-6">
        {/* Language */}
        <div className="rounded-xl border border-slate-700 bg-slate-800/50 p-5">
          <h3 className="text-lg font-bold mb-4">Language</h3>
          <div className="grid grid-cols-3 gap-3">
            {languages.map(lang => (
              <button
                key={lang.code}
                onClick={() => onLanguageChange(lang.code)}
                className={`rounded-lg px-4 py-3 text-sm transition-colors ${
                  language === lang.code
                    ? 'bg-nexus-500/20 text-nexus-400 border border-nexus-500/50'
                    : 'bg-slate-700 text-slate-400 hover:bg-slate-600'
                }`}
              >
                {lang.name}
              </button>
            ))}
          </div>
        </div>

        {/* Accessibility */}
        <div className="rounded-xl border border-slate-700 bg-slate-800/50 p-5">
          <h3 className="text-lg font-bold mb-4">Accessibility</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Large Text</p>
                <p className="text-xs text-slate-500">Increase text size for better readability</p>
              </div>
              <ToggleSwitch checked={largeText} onChange={onLargeTextChange} />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">High Contrast</p>
                <p className="text-xs text-slate-500">Enhanced contrast for visibility</p>
              </div>
              <ToggleSwitch checked={highContrast} onChange={onHighContrastChange} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function YouTubeScreen({ onBack }: { onBack: () => void }) {
  const [searchQuery, setSearchQuery] = useState('');
  const [videos, setVideos] = useState<Array<{id: string; title: string; description: string; thumbnail: string; channel: string; duration: string; video_id: string}>>([
    { id: '1', title: 'Welcome to Grand Plaza Hotel', description: 'Experience luxury at Grand Plaza Hotel', thumbnail: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg', channel: 'Grand Plaza Official', duration: '2:30', video_id: 'dQw4w9WgXcQ' },
    { id: '2', title: 'Local Attractions Guide', description: 'Discover the best places near our hotel', thumbnail: 'https://i.ytimg.com/vi/9bZkp7q19f0/hqdefault.jpg', channel: 'Travel Guide', duration: '5:45', video_id: '9bZkp7q19f0' },
    { id: '3', title: 'Spa & Wellness Introduction', description: 'Relax and rejuvenate at our spa', thumbnail: 'https://i.ytimg.com/vi/kJQP7kiw5Fk/hqdefault.jpg', channel: 'Grand Plaza Spa', duration: '3:15', video_id: 'kJQP7kiw5Fk' },
  ]);
  const [selectedVideoId, setSelectedVideoId] = useState<string | null>(null);

  const trendingVideos = [
    { id: 't1', title: 'Top 10 Luxury Hotels 2026', thumbnail: 'https://i.ytimg.com/vi/M7lc1UVf-VE/hqdefault.jpg', channel: 'Travel + Leisure', views: '1.2M', video_id: 'M7lc1UVf-VE' },
    { id: 't2', title: 'Best Room Service Experiences', thumbnail: 'https://i.ytimg.com/vi/9bZkp7q19f0/hqdefault.jpg', channel: 'Hotel Management', views: '856K', video_id: '9bZkp7q19f0' },
    { id: 't3', title: 'Hotel Technology Trends', thumbnail: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg', channel: 'Hospitality Tech', views: '430K', video_id: 'dQw4w9WgXcQ' },
  ];

  return (
    <div className="p-8">
      <div className="flex items-center gap-4 mb-6">
        <button onClick={onBack} className="rounded-lg p-2 hover:bg-slate-800">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h2 className="text-2xl font-bold">YouTube</h2>
      </div>

      {/* Search */}
      <div className="relative mb-6">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          placeholder="Search videos..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="input w-full pl-10"
        />
      </div>

      {/* Selected Video Player */}
      {selectedVideoId && (
        <div className="mb-6 rounded-xl border border-slate-700 bg-slate-800/50 p-4">
          <div className="aspect-video rounded-lg bg-slate-900 overflow-hidden">
            <iframe
              src={`https://www.youtube.com/embed/${selectedVideoId}?autoplay=1&rel=0&modestbranding=1`}
              className="w-full h-full"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
            />
          </div>
          <button 
            onClick={() => setSelectedVideoId(null)}
            className="mt-3 text-sm text-slate-400 hover:text-white"
          >
            Close Player
          </button>
        </div>
      )}

      {/* Hotel Videos */}
      <h3 className="text-lg font-bold mb-4">Hotel Content</h3>
      <div className="grid grid-cols-3 gap-4 mb-8">
        {videos.map(video => (
          <button
            key={video.id}
            onClick={() => setSelectedVideoId(video.video_id)}
            className="group rounded-xl border border-slate-700 overflow-hidden text-left hover:border-rose-500/50 transition-all"
          >
            <div className="h-32 bg-slate-700/50 flex items-center justify-center relative">
              <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                <div className="rounded-full bg-rose-500/20 p-3">
                  <Play className="h-6 w-6 text-rose-400" />
                </div>
              </div>
              <span className="absolute bottom-2 right-2 rounded bg-slate-900/80 px-1.5 py-0.5 text-[10px] text-white">
                {video.duration}
              </span>
            </div>
            <div className="p-3">
              <p className="font-medium text-sm line-clamp-2">{video.title}</p>
              <p className="text-xs text-slate-500 mt-1">{video.channel}</p>
            </div>
          </button>
        ))}
      </div>

      {/* Trending */}
      <h3 className="text-lg font-bold mb-4">Trending</h3>
      <div className="grid grid-cols-3 gap-4">
        {trendingVideos.map(video => (
          <button
            key={video.id}
            onClick={() => setSelectedVideoId(video.video_id)}
            className="group rounded-xl border border-slate-700 overflow-hidden text-left hover:border-rose-500/50 transition-all"
          >
            <div className="h-32 bg-slate-700/50 flex items-center justify-center relative">
              <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                <div className="rounded-full bg-rose-500/20 p-3">
                  <Play className="h-6 w-6 text-rose-400" />
                </div>
              </div>
            </div>
            <div className="p-3">
              <p className="font-medium text-sm line-clamp-2">{video.title}</p>
              <p className="text-xs text-slate-500 mt-1">{video.channel} · {video.views} views</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

function ToggleSwitch({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button 
      onClick={() => onChange(!checked)}
      className={`relative inline-flex h-7 w-12 items-center rounded-full transition-colors ${
        checked ? 'bg-nexus-500' : 'bg-slate-700'
      }`}
    >
      <span className={`inline-block h-5 w-5 transform rounded-full bg-white transition-transform ${
        checked ? 'translate-x-6' : 'translate-x-1'
      }`} />
    </button>
  );
}
