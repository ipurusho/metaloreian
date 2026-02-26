import { Link } from 'react-router-dom';

interface BandLinkProps {
  bandId: number;
  bandName: string;
  className?: string;
}

export function BandLink({ bandId, bandName, className }: BandLinkProps) {
  return (
    <Link to={`/band/${bandId}`} className={className || 'band-link'}>
      {bandName}
    </Link>
  );
}
