export function formattedDate(dateStr: string): string {
    let date = new Date(dateStr);
    let formattedDate = new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'long', day: '2-digit' }).format(date);
    return formattedDate
}
