'use client';

import { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { uploadReceiptScan, fetchReceiptAudits } from '../services/api';
import { ReceiptAudit } from '../types';
import {
  UploadCloud,
  FileText,
  Sparkles,
  Receipt,
  CheckCircle2,
  AlertCircle,
  RefreshCw,
  Calendar,
  Store,
  DollarSign,
} from 'lucide-react';

export function ReceiptUploader() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [scannedAudit, setScannedAudit] = useState<ReceiptAudit | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Fetch previous receipt audit history
  const { data: auditHistory = [], refetch: refetchHistory } = useQuery({
    queryKey: ['receiptAudits'],
    queryFn: fetchReceiptAudits,
  });

  // OCR Scan Mutation
  const scanMutation = useMutation({
    mutationFn: uploadReceiptScan,
    onSuccess: (data) => {
      setScannedAudit(data);
      setErrorMsg(null);
      refetchHistory();
    },
    onError: (err: any) => {
      const msg = err.response?.data?.error || err.message || 'AI Receipt scanning failed';
      setErrorMsg(msg);
    },
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setSelectedFile(file);
      setPreviewUrl(URL.createObjectURL(file));
      setScannedAudit(null);
      setErrorMsg(null);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const file = e.dataTransfer.files[0];
      setSelectedFile(file);
      setPreviewUrl(URL.createObjectURL(file));
      setScannedAudit(null);
      setErrorMsg(null);
    }
  };

  const handleUpload = () => {
    if (!selectedFile) return;
    scanMutation.mutate(selectedFile);
  };

  return (
    <div className="space-y-8">
      {/* Top Banner Header */}
      <div className="glass-panel p-6 rounded-2xl border border-brand-500/30 flex flex-wrap items-center justify-between gap-4">
        <div>
          <div className="flex items-center space-x-2 text-brand-500 font-bold text-xs uppercase tracking-widest mb-1">
            <Sparkles className="w-4 h-4 animate-pulse" />
            <span>Smart AI Vision Engine</span>
          </div>
          <h2 className="text-xl font-black text-white">Automated Receipt Expense Auditor</h2>
          <p className="text-xs text-gray-400 mt-1">
            Upload paper receipt images to extract merchant names, line items, and prices automatically using Gemini Vision AI.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Drag & Drop Upload Zone (5 Cols) */}
        <div className="lg:col-span-5 glass-panel p-6 rounded-2xl border border-gray-700/50 flex flex-col justify-between">
          <div>
            <h3 className="text-sm font-bold text-gray-200 mb-4 flex items-center space-x-2">
              <UploadCloud className="w-4 h-4 text-brand-500" />
              <span>Select Paper Receipt Image</span>
            </h3>

            {/* Drag Drop Box */}
            <div
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleDrop}
              className="border-2 border-dashed border-gray-700 hover:border-brand-500/60 rounded-2xl p-6 text-center bg-dark-800/60 transition-colors cursor-pointer relative"
            >
              <input
                type="file"
                accept="image/*"
                onChange={handleFileChange}
                className="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
              />

              {previewUrl ? (
                <div className="space-y-3">
                  <img
                    src={previewUrl}
                    alt="Receipt Preview"
                    className="max-h-48 rounded-xl mx-auto object-contain border border-gray-700/50 shadow-lg"
                  />
                  <span className="text-xs font-semibold text-gray-300 block truncate">
                    {selectedFile?.name}
                  </span>
                </div>
              ) : (
                <div className="py-8 space-y-3">
                  <div className="w-12 h-12 rounded-full bg-brand-500/10 text-brand-500 flex items-center justify-center mx-auto">
                    <Receipt className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-gray-200">
                      Drag & drop receipt image here
                    </p>
                    <p className="text-[10px] text-gray-400 mt-1">
                      Supports JPG, PNG, WEBP (Max 10MB)
                    </p>
                  </div>
                </div>
              )}
            </div>

            {errorMsg && (
              <div className="mt-4 bg-red-500/10 border border-red-500/30 text-red-400 p-3 rounded-xl text-xs flex items-center space-x-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{errorMsg}</span>
              </div>
            )}
          </div>

          <button
            onClick={handleUpload}
            disabled={!selectedFile || scanMutation.isPending}
            className={`w-full mt-6 py-3 rounded-xl font-bold text-xs flex items-center justify-center space-x-2 transition-all ${
              !selectedFile || scanMutation.isPending
                ? 'bg-gray-800 text-gray-500 cursor-not-allowed'
                : 'bg-brand-500 text-black hover:bg-brand-600 shadow-lg shadow-brand-500/20 active:scale-[0.98]'
            }`}
          >
            {scanMutation.isPending ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin text-black" />
                <span>Analyzing Receipt with AI...</span>
              </>
            ) : (
              <>
                <Sparkles className="w-4 h-4 text-black" />
                <span>Scan Receipt Expenses</span>
              </>
            )}
          </button>
        </div>

        {/* Right Extracted OCR Preview Card (7 Cols) */}
        <div className="lg:col-span-7 glass-panel p-6 rounded-2xl border border-gray-700/50 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-4 border-b border-gray-700/50 mb-4">
              <h3 className="text-sm font-bold text-gray-200 flex items-center space-x-2">
                <FileText className="w-4 h-4 text-brand-500" />
                <span>Extracted Expense Breakdown</span>
              </h3>

              {scannedAudit && (
                <span className="text-[10px] font-bold text-brand-500 bg-brand-500/10 px-2.5 py-1 rounded-lg border border-brand-500/20 flex items-center space-x-1">
                  <CheckCircle2 className="w-3 h-3" />
                  <span>AI Audit Verified</span>
                </span>
              )}
            </div>

            {!scannedAudit ? (
              <div className="py-16 text-center text-gray-500 flex flex-col items-center">
                <Receipt className="w-12 h-12 text-gray-600 stroke-[1.5] mb-2" />
                <p className="text-sm font-medium">No active receipt scanned</p>
                <p className="text-xs text-gray-500 mt-1">
                  Upload a paper receipt to view extracted JSON items and pricing
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {/* Merchant & Date Meta */}
                <div className="grid grid-cols-2 gap-3 bg-dark-800/80 p-3.5 rounded-xl border border-gray-700/50 text-xs">
                  <div>
                    <span className="text-gray-400 text-[10px] block flex items-center space-x-1">
                      <Store className="w-3 h-3 text-gray-400" />
                      <span>Merchant Name</span>
                    </span>
                    <strong className="text-white font-bold mt-0.5 block">
                      {scannedAudit.merchant_name}
                    </strong>
                  </div>
                  <div>
                    <span className="text-gray-400 text-[10px] block flex items-center space-x-1">
                      <Calendar className="w-3 h-3 text-gray-400" />
                      <span>Receipt Date</span>
                    </span>
                    <strong className="text-gray-200 font-bold mt-0.5 block">
                      {new Date(scannedAudit.receipt_date).toLocaleDateString()}
                    </strong>
                  </div>
                </div>

                {/* Items Table */}
                <div className="bg-dark-800/60 rounded-xl border border-gray-700/50 overflow-hidden">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-dark-700/50 text-gray-400 text-[10px] uppercase font-bold border-b border-gray-700/50">
                      <tr>
                        <th className="p-3">Item Description</th>
                        <th className="p-3 text-center">Qty</th>
                        <th className="p-3 text-right">Price</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-700/40 text-gray-200">
                      {scannedAudit.raw_ocr_json.items.map((item, idx) => (
                        <tr key={idx} className="hover:bg-dark-700/30">
                          <td className="p-3 font-semibold">{item.name}</td>
                          <td className="p-3 text-center font-mono">{item.quantity}</td>
                          <td className="p-3 text-right font-mono">${item.price.toFixed(2)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Total Extracted Sum */}
                <div className="flex justify-between items-center bg-brand-500/10 border border-brand-500/30 p-3.5 rounded-xl">
                  <span className="text-xs font-bold text-gray-200 flex items-center space-x-1">
                    <DollarSign className="w-4 h-4 text-brand-500" />
                    <span>Total Expense Amount</span>
                  </span>
                  <span className="text-lg font-black text-brand-500">
                    ${scannedAudit.total_amount.toFixed(2)}
                  </span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Audit History Log Table */}
      <div className="glass-panel p-6 rounded-2xl border border-gray-700/50">
        <h3 className="text-sm font-bold text-gray-200 mb-4 flex items-center space-x-2">
          <FileText className="w-4 h-4 text-brand-500" />
          <span>Receipt Audit History Log</span>
        </h3>

        {auditHistory.length === 0 ? (
          <p className="text-xs text-gray-500 py-4 text-center">
            No previous receipt audits logged in PostgreSQL database.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-dark-800 text-gray-400 text-[10px] uppercase font-bold border-b border-gray-700/50">
                <tr>
                  <th className="p-3">Audit ID</th>
                  <th className="p-3">Merchant</th>
                  <th className="p-3">Receipt Date</th>
                  <th className="p-3 text-right">Total Amount</th>
                  <th className="p-3 text-right">Created At</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/40 text-gray-300">
                {auditHistory.map((audit) => (
                  <tr key={audit.id} className="hover:bg-dark-800/50">
                    <td className="p-3 font-mono text-[10px] text-gray-400">{audit.id}</td>
                    <td className="p-3 font-semibold text-white">{audit.merchant_name}</td>
                    <td className="p-3">
                      {new Date(audit.receipt_date).toLocaleDateString()}
                    </td>
                    <td className="p-3 text-right font-bold text-brand-500">
                      ${audit.total_amount.toFixed(2)}
                    </td>
                    <td className="p-3 text-right text-[10px] text-gray-400">
                      {new Date(audit.created_at).toLocaleTimeString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
