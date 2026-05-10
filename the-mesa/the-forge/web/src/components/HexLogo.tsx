export default function HexLogo({ collapsed }: { collapsed: boolean }) {
  return (
    <svg
      width="36"
      height="36"
      viewBox="0 0 36 36"
      className={collapsed ? 'mx-auto' : 'mx-auto mb-2'}
    >
      <polygon
        points="18,2 32,10 32,26 18,34 4,26 4,10"
        fill="none"
        stroke="hsl(var(--primary))"
        strokeWidth="1"
        strokeOpacity="0.4"
        className="origin-center"
        style={{ animation: 'spin 60s linear infinite', transformOrigin: '18px 18px' }}
      />
      <polygon
        points="18,7 27,12.5 27,23.5 18,29 9,23.5 9,12.5"
        fill="hsl(var(--primary))"
        fillOpacity="0.12"
        stroke="hsl(var(--primary))"
        strokeWidth="0.8"
        style={{ animation: 'pulse-opacity 3s ease-in-out infinite' }}
      />
      <polygon
        points="18,11 23,14 23,22 18,25 13,22 13,14"
        fill="hsl(var(--primary))"
        fillOpacity="0.06"
        stroke="hsl(var(--primary))"
        strokeWidth="0.4"
      />
      <text
        x="18"
        y="21"
        textAnchor="middle"
        fill="hsl(var(--primary))"
        fontSize="8"
        fontFamily="monospace"
      >
        F
      </text>
    </svg>
  );
}
