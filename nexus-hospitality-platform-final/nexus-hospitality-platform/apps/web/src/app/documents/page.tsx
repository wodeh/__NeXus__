'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';

// Document Manager — upload, view, download files

export default function DocumentsPage() {
  const [selectedType, setSelectedType] = useState('all');

  const documentTypes = [
    { id: 'all', label: 'All Documents' },
    { id: 'guest_id', label: 'Guest IDs' },
    { id: 'contract', label: 'Contracts' },
    { id: 'invoice', label: 'Invoices' },
    { id: 'receipt', label: 'Receipts' },
    { id: 'housekeeping', label: 'Housekeeping' },
    { id: 'maintenance', label: 'Maintenance' },
  ];

  const documents = [
    { name: 'Passport_JohnSmith.pdf', type: 'guest_id', entity: 'John Smith', size: '2.4 MB', date: 'May 10, 2026', uploadedBy: 'Alice' },
    { name: 'Booking_Contract_CorpLLC.pdf', type: 'contract', entity: 'Corporate LLC', size: '1.1 MB', date: 'May 8, 2026', uploadedBy: 'Bob' },
    { name: 'Invoice_2026_001.pdf', type: 'invoice', entity: 'John Smith', size: '145 KB', date: 'May 10, 2026', uploadedBy: 'System' },
    { name: 'Room_201_Cleaning_Photo.jpg', type: 'housekeeping', entity: 'Room 201', size: '3.2 MB', date: 'May 9, 2026', uploadedBy: 'Carol' },
    { name: 'AC_Repair_Receipt.pdf', type: 'maintenance', entity: 'Room 305', size: '280 KB', date: 'May 7, 2026', uploadedBy: 'Dave' },
    { name: 'ID_MariaGarcia.jpg', type: 'guest_id', entity: 'Maria Garcia', size: '1.8 MB', date: 'May 9, 2026', uploadedBy: 'Alice' },
  ];

  const filteredDocs = selectedType === 'all' 
    ? documents 
    : documents.filter(d => d.type === selectedType);

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'guest_id': return 'bg-blue-500/20 text-blue-400';
      case 'contract': return 'bg-purple-500/20 text-purple-400';
      case 'invoice': return 'bg-emerald-500/20 text-emerald-400';
      case 'receipt': return 'bg-amber-500/20 text-amber-400';
      case 'housekeeping': return 'bg-cyan-500/20 text-cyan-400';
      case 'maintenance': return 'bg-rose-500/20 text-rose-400';
      default: return 'bg-slate-500/20 text-slate-400';
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'guest_id': return 'Guest ID';
      case 'contract': return 'Contract';
      case 'invoice': return 'Invoice';
      case 'receipt': return 'Receipt';
      case 'housekeeping': return 'Housekeeping';
      case 'maintenance': return 'Maintenance';
      default: return type;
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-white p-6">
      <div className="max-w-7xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold">Documents</h1>
          <p className="text-slate-400 mt-1">Manage files and attachments</p>
        </header>

        {/* Upload Area */}
        <Card className="mb-6 bg-slate-800/50 border-slate-700">
          <CardContent className="pt-6">
            <div className="border-2 border-dashed border-slate-600 rounded-lg p-8 text-center hover:border-amber-500/50 transition-colors">
              <div className="text-4xl mb-3">📁</div>
              <p className="text-slate-400 mb-2">Drag and drop files here, or click to browse</p>
              <Button className="bg-amber-500 hover:bg-amber-600 text-slate-900 font-semibold">
                Upload File
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Filters */}
        <div className="flex gap-2 mb-6 flex-wrap">
          {documentTypes.map((type) => (
            <Button
              key={type.id}
              variant={selectedType === type.id ? 'default' : 'outline'}
              size="sm"
              className={selectedType === type.id ? 'bg-amber-500 text-slate-900' : 'border-slate-600 text-slate-400'}
              onClick={() => setSelectedType(type.id)}
            >
              {type.label}
            </Button>
          ))}
        </div>

        {/* Document List */}
        <Card className="bg-slate-800/50 border-slate-700">
          <CardHeader>
            <CardTitle>All Files</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {filteredDocs.map((doc, i) => (
                <div key={i} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-lg bg-slate-600 flex items-center justify-center text-2xl">
                      📄
                    </div>
                    <div>
                      <p className="font-medium text-white">{doc.name}</p>
                      <p className="text-sm text-slate-400">
                        {doc.entity} • {doc.size} • {doc.date}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge className={getTypeColor(doc.type)}>
                      {getTypeLabel(doc.type)}
                    </Badge>
                    <span className="text-sm text-slate-500">by {doc.uploadedBy}</span>
                    <Button size="sm" variant="outline" className="border-slate-600">
                      Download
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
