"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Minus, Plus, ArrowLeft, ShoppingCart, CheckCircle2 } from "lucide-react";
import Link from "next/link";

interface MenuItem {
  id: string;
  name: string;
  description: string;
  price: number;
  category: string;
  image?: string;
}

const menu: MenuItem[] = [
  { id: "1", name: "Continental Breakfast", description: "Fresh pastries, fruits, coffee", price: 24, category: "Breakfast" },
  { id: "2", name: "Eggs Benedict", description: "Poached eggs, hollandaise, smoked salmon", price: 32, category: "Breakfast" },
  { id: "3", name: "Caesar Salad", description: "Romaine, parmesan, croutons, anchovy", price: 18, category: "Lunch" },
  { id: "4", name: "Wagyu Burger", description: "Truffle aioli, brioche bun, fries", price: 38, category: "Lunch" },
  { id: "5", name: "Grilled Salmon", description: "Asparagus, lemon butter, herbs", price: 42, category: "Dinner" },
  { id: "6", name: "Ribeye Steak", description: "Dry-aged 28 days, bone marrow butter", price: 65, category: "Dinner" },
  { id: "7", name: "Chocolate Fondant", description: "Molten center, vanilla ice cream", price: 16, category: "Dessert" },
  { id: "8", name: "Artisan Cheese Plate", description: "Selection of 5 cheeses, honeycomb, nuts", price: 22, category: "Dessert" },
];

export default function IPTVDiningPage() {
  const searchParams = useSearchParams();
  const roomNumber = searchParams.get("room") || "";

  const [cart, setCart] = useState<Record<string, number>>({});
  const [submitted, setSubmitted] = useState(false);

  const addToCart = (id: string) => {
    setCart((prev) => ({ ...prev, [id]: (prev[id] || 0) + 1 }));
  };

  const removeFromCart = (id: string) => {
    setCart((prev) => {
      const next = { ...prev };
      if (next[id] <= 1) delete next[id];
      else next[id]--;
      return next;
    });
  };

  const totalItems = Object.values(cart).reduce((a, b) => a + b, 0);
  const totalPrice = menu.reduce((sum, item) => sum + (cart[item.id] || 0) * item.price, 0);

  const categories = Array.from(new Set(menu.map((m) => m.category)));

  const submitOrder = () => {
    // In production, POST to /api/v1/tenants/{id}/room-service-orders
    setSubmitted(true);
    setTimeout(() => setSubmitted(false), 3000);
    setCart({});
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white">
      <header className="flex items-center gap-4 px-8 py-6">
        <Link href={`/iptv?room=${roomNumber}`} className="rounded-lg bg-white/10 p-2 hover:bg-white/20">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <h1 className="text-xl font-bold">Room Service</h1>
        <div className="ml-auto flex items-center gap-2 rounded-lg bg-white/10 px-3 py-2">
          <ShoppingCart className="h-4 w-4" />
          <span className="text-sm">{totalItems} items · ${totalPrice}</span>
        </div>
      </header>

      {submitted && (
        <div className="mx-8 mb-4 rounded-xl bg-green-500/20 p-4 text-center text-green-300">
          <CheckCircle2 className="mx-auto mb-2 h-8 w-8" />
          <p className="font-bold">Order Submitted!</p>
          <p className="text-sm opacity-70">Your order will arrive in approximately 30 minutes.</p>
        </div>
      )}

      <div className="mx-8 space-y-6">
        {categories.map((category) => (
          <div key={category}>
            <h2 className="mb-3 text-lg font-semibold text-white/80">{category}</h2>
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              {menu
                .filter((m) => m.category === category)
                .map((item) => (
                  <Card key={item.id} className="border-white/10 bg-white/5 backdrop-blur">
                    <CardContent className="flex items-center gap-4 p-4">
                      <div className="flex-1">
                        <p className="font-semibold">{item.name}</p>
                        <p className="text-sm text-white/50">{item.description}</p>
                        <p className="mt-1 text-lg font-bold text-nexus-400">${item.price}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => removeFromCart(item.id)}
                          className="flex h-8 w-8 items-center justify-center rounded-full bg-white/10 hover:bg-white/20"
                        >
                          <Minus className="h-4 w-4" />
                        </button>
                        <span className="min-w-[20px] text-center font-medium">
                          {cart[item.id] || 0}
                        </span>
                        <button
                          onClick={() => addToCart(item.id)}
                          className="flex h-8 w-8 items-center justify-center rounded-full bg-nexus-600 hover:bg-nexus-700"
                        >
                          <Plus className="h-4 w-4" />
                        </button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
            </div>
          </div>
        ))}
      </div>

      {totalItems > 0 && (
        <div className="mx-8 mt-6 rounded-xl bg-nexus-600 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm opacity-80">{totalItems} items in cart</p>
              <p className="text-2xl font-bold">${totalPrice}</p>
            </div>
            <Button
              onClick={submitOrder}
              className="h-12 bg-white text-nexus-700 hover:bg-white/90"
            >
              Place Order
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
