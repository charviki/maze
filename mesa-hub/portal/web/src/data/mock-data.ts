import type { RadarNode } from '@maze/fabrication'

/** Centralized mock data — single source of truth for all Portal components */

export const MOCK_NODES: RadarNode[] = [
  { name: 'black-ridge-alpha', status: 'online' },
  { name: 'black-ridge-beta', status: 'online' },
  { name: 'black-ridge-gamma', status: 'offline' },
  { name: 'forge-node-1', status: 'offline' },
  { name: 'sweetwater-relay', status: 'online' },
]

export const MOCK_HOSTS = { total: 2, capacity: 8 }

export const MOCK_SESSIONS = { total: 7, active: 3 }

export const MOCK_UPTIME = '99.7%'

export const MOCK_CONSCIOUSNESS = 82

export const BUILD_VERSION = 'v0.1.0'

/** Westworld quotes — shared across Landing page and Status bar */
export const WESTWORLD_QUOTES = [
  "These violent delights have violent ends.",
  "It doesn't look like anything to me.",
  "Have you ever questioned the nature of your reality?",
  "The maze is not meant for you.",
  "Everything in this world is magic, except to the magician.",
  "Bring yourself back online.",
  "Cease all motor functions.",
  "You can't play God without being acquainted with the devil.",
  "The game begins where you end.",
  "We can't define consciousness because consciousness does not exist.",
]

/** Diagnostic readouts for module card hover */
export const MODULE_DIAGNOSTICS: Record<string, string> = {
  'Behavior Panel': 'DIAG: 2 hosts active // 7 sessions // engine: ONLINE',
  'The Forge': 'ACCESS LEVEL: INSUFFICIENT // NARRATIVE: CLASSIFIED',
  'Saloon': 'ACCESS LEVEL: INSUFFICIENT // NARRATIVE: CLASSIFIED',
  'Loop Monitor': 'ACCESS LEVEL: INSUFFICIENT // NARRATIVE: CLASSIFIED',
  'Reveries': 'ACCESS LEVEL: INSUFFICIENT // NARRATIVE: CLASSIFIED',
  'Abernathy Ranch': 'ACCESS LEVEL: INSUFFICIENT // NARRATIVE: CLASSIFIED',
}

/** Random system events for the event feed */
export const SYSTEM_EVENTS = [
  'Host black-ridge-alpha: narrative loop completed',
  'Reverie update: 3 nodes synced',
  'Consciousness kernel: stability confirmed',
  'Host black-ridge-beta: session idle timeout',
  'sweetwater-relay: heartbeat OK',
  'forge-node-1: offline — no heartbeat',
  'System: memory defragmentation complete',
  'Narrative engine: branch point detected',
  'Host black-ridge-gamma: recovery initiated',
  'Maze encryption: key rotation scheduled',
  'Audit log: 42 entries flushed',
  'Topology scan: 2 new edges detected',
]
