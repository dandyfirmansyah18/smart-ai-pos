export const formatIDR = (amount: number): string => {
  if (isNaN(amount)) return 'Rp. 0';
  const rounded = Math.round(amount);
  const formatted = rounded.toLocaleString('id-ID');
  return `Rp. ${formatted}`;
};
