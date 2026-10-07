export interface SessionPosture {
  profile: string;
  unenforcedControls: string[];
}
export function Posture({ posture }: { posture: SessionPosture }) {
  return (
    <section aria-labelledby="posture-title">
      <h2 id="posture-title">Effective posture</h2>
      <dl>
        <dt>Profile</dt>
        <dd>{posture.profile}</dd>
        <dt>Not enforced</dt>
        <dd>
          <ul>
            {posture.unenforcedControls.map((control) => (
              <li key={control}>{control}</li>
            ))}
          </ul>
        </dd>
      </dl>
      <p className="warning">Network egress is not enforced by this profile.</p>
    </section>
  );
}
