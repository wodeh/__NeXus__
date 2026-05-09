// clients/web-portal/src/i18n/index.ts
// ============================================================
// MULTILINGUAL CONCIERGE AI — 40+ languages
// RTL support, hospitality terminology, regional formatting
// ============================================================

import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import ICU from 'i18next-icu';
import LanguageDetector from 'i18next-browser-languagedetector';
import HttpApi from 'i18next-http-backend';

// Core languages (hospitality priority)
const coreLanguages = [
  'en', 'es', 'fr', 'de', 'it', 'pt', 'ru', 'zh', 'ja', 'ko',
  'ar', 'he', 'hi', 'th', 'vi', 'tr', 'nl', 'pl', 'sv', 'da',
  'no', 'fi', 'el', 'cs', 'hu', 'ro', 'id', 'ms', 'uk', 'bg',
  'hr', 'sr', 'sk', 'sl', 'et', 'lv', 'lt', 'mt', 'is', 'ga',
];

// Hospitality-specific translation namespaces
const namespaces = [
  'common',           // General UI
  'guest',            // Guest-facing
  'staff',            // Staff operations
  'housekeeping',     // Room cleaning
  'maintenance',      // Repairs
  'reservations',     // Booking
  'billing',          // Payments
  'iptv',             // TV interface
  'iot',              // Smart room
  'dining',           // F&B
  'spa',              // Wellness
  'concierge',        // Services
  'emergency',        // Safety
  'loyalty',          // Rewards
  'analytics',        // Reports
];

i18n
  .use(HttpApi)
  .use(ICU)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en',
    supportedLngs: coreLanguages,
    ns: namespaces,
    defaultNS: 'common',
    interpolation: {
      escapeValue: false, // React handles escaping
    },
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'nexus-language',
      caches: ['localStorage'],
    },
    backend: {
      loadPath: '/api/i18n/{{lng}}/{{ns}}',
    },
    react: {
      useSuspense: true,
      transSupportBasicHtmlNodes: true,
      transKeepBasicHtmlNodesFor: ['br', 'strong', 'i', 'p'],
    },
  });

// RTL language support
const rtlLanguages = ['ar', 'he', 'fa', 'ur'];

export const isRTL = (language: string): boolean => {
  return rtlLanguages.includes(language);
};

export const getTextDirection = (language: string): 'ltr' | 'rtl' => {
  return isRTL(language) ? 'rtl' : 'ltr';
};

// Hospitality-specific formatting
export const formatCurrency = (amount: number, currency: string, locale: string): string => {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
  }).format(amount);
};

export const formatDate = (date: Date, locale: string, options?: Intl.DateTimeFormatOptions): string => {
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    ...options,
  }).format(date);
};

export const formatRelativeTime = (date: Date, locale: string): string => {
  const now = new Date();
  const diff = date.getTime() - now.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });

  if (Math.abs(minutes) < 60) return rtf.format(minutes, 'minute');
  if (Math.abs(hours) < 24) return rtf.format(hours, 'hour');
  return rtf.format(days, 'day');
};

// Room number formatting (varies by region)
export const formatRoomNumber = (roomNumber: string, locale: string): string => {
  // Some regions use floor-room format (e.g., 1205 = 12th floor, room 05)
  if (locale.startsWith('zh') || locale.startsWith('ja') || locale.startsWith('ko')) {
    return roomNumber; // No modification for East Asian
  }
  // Western: may add floor prefix
  return roomNumber;
};

export default i18n;
