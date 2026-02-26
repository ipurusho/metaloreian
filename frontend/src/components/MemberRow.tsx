import type { Member } from '../api/client';
import { BandLink } from './BandLink';

interface MemberRowProps {
  member: Member;
}

export function MemberRow({ member }: MemberRowProps) {
  return (
    <div className="member-row">
      <div className="member-info">
        <span className="member-name">{member.name}</span>
        <span className="member-instrument">{member.instrument}</span>
      </div>
      {member.other_bands && member.other_bands.length > 0 && (
        <div className="member-other-bands">
          <span className="other-bands-label">See also: </span>
          {member.other_bands.map((band, i) => (
            <span key={band.band_id}>
              {i > 0 && ', '}
              <BandLink bandId={band.band_id} bandName={band.band_name} className="other-band-link" />
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
