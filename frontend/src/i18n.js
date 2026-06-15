// MiyooDeck - Internationalization (FR/EN)

const translations = {
  fr: {
    // App
    appName: 'MiyooDeck',
    appTagline: 'Ton Miyoo dans ton navigateur',
    connecting: 'Connexion...',
    gameRunning: 'En jeu',

    // Connection error
    connErrTitle: 'Connexion impossible',
    connErrMsg: 'MiyooDeck ne répond pas sur',
    connErrStep1: 'Assure-toi que la Miyoo est allumée',
    connErrStep2: 'Lance Apps → Web Deck sur la console',
    connErrStep3: 'Vérifie que le WiFi est activé',
    connErrStep4: 'Recharge cette page',
    connErrRetry: '🔄 Réessayer',

    // Nav
    navDashboard: 'Tableau de bord',
    navGames: 'Jeux',
    navFiles: 'Fichiers',
    navConfig: 'Config',

    // Auth
    loginTitle: 'Connexion',
    loginSubtitle: 'Entrez votre code PIN pour vous connecter',
    loginPin: 'Code PIN',
    loginButton: 'Se connecter',
    loginConnecting: 'Connexion...',
    setupTitle: 'Première connexion',
    setupSubtitle: 'Définissez un code PIN pour protéger votre console',
    setupNewPin: 'Nouveau PIN (min. 4 chiffres)',
    setupConfirm: 'Confirmer le PIN',
    setupButton: 'Créer le PIN et entrer',
    setupSkip: 'Ignorer (sans protection)',
    pinMismatch: 'Les PINs ne correspondent pas',
    pinTooShort: 'Le PIN doit contenir au moins 4 chiffres',
    wrongPin: 'Code PIN incorrect',

    // Dashboard
    cpu: 'Processeur',
    ram: 'Mémoire',
    battery: 'Batterie',
    network: 'Réseau',
    uptime: 'Uptime',
    liveScreen: 'Écran en direct',
    refresh: 'Rafraîchir',
    savePng: '⬇ Capture PNG',
    screenUnavailable: 'Aperçu non disponible',
    startGame: 'Lancez un jeu pour voir l\'écran',

    // Games
    systems: 'Systèmes',
    searchRoms: 'Rechercher des ROMs...',
    selectSystem: 'Sélectionnez un système',
    selectSystemSub: 'pour parcourir les ROMs',
    noRoms: 'Aucune ROM dans ce système',
    launching: 'Lancement...',
    launchError: 'Erreur de lancement',
    noSystemsFound: 'Aucune ROM trouvée sur la carte SD',

    // Files
    upload: '⬆ Envoyer',
    savesBackup: '⬇ Sauvegardes',
    dropToUpload: 'Déposez les fichiers ici\npour les envoyer vers',
    deleteConfirm: (name) => `Supprimer "${name}" ?`,
    deleteSuccess: 'Supprimé : ',
    uploadComplete: 'Envoi terminé',
    extract: 'Extraire',
    emptyDir: 'Dossier vide',

    // Config
    configFiles: 'Fichiers de config',
    selectConfig: 'Sélectionnez un fichier',
    backupNote: 'Sauvegarde automatique avant chaque modification',
    save: 'Sauvegarder (Ctrl+S)',
    saving: 'Sauvegarde...',
    unsaved: 'Modifications non sauvegardées',
    savedMsg: '✓ Sauvegardé',
    notEditable: 'Ce type de fichier ne peut pas être édité',
  },

  en: {
    // App
    appName: 'MiyooDeck',
    appTagline: 'Your Miyoo in your browser',
    connecting: 'Connecting...',
    gameRunning: 'In game',

    // Connection error
    connErrTitle: 'Connection failed',
    connErrMsg: 'MiyooDeck is not responding on',
    connErrStep1: 'Make sure the Miyoo is powered on',
    connErrStep2: 'Open Apps → Web Deck on the console',
    connErrStep3: 'Check that WiFi is enabled',
    connErrStep4: 'Reload this page',
    connErrRetry: '🔄 Retry',

    // Nav
    navDashboard: 'Dashboard',
    navGames: 'Games',
    navFiles: 'Files',
    navConfig: 'Config',

    // Auth
    loginTitle: 'Login',
    loginSubtitle: 'Enter your PIN to connect',
    loginPin: 'PIN code',
    loginButton: 'Connect',
    loginConnecting: 'Connecting...',
    setupTitle: 'First connection',
    setupSubtitle: 'Set a PIN to protect your console',
    setupNewPin: 'New PIN (min. 4 digits)',
    setupConfirm: 'Confirm PIN',
    setupButton: 'Create PIN & Enter',
    setupSkip: 'Skip (no protection)',
    pinMismatch: 'PINs do not match',
    pinTooShort: 'PIN must be at least 4 digits',
    wrongPin: 'Wrong PIN',

    // Dashboard
    cpu: 'CPU',
    ram: 'RAM',
    battery: 'Battery',
    network: 'Network',
    uptime: 'Uptime',
    liveScreen: 'Live Screen',
    refresh: 'Refresh',
    savePng: '⬇ Save PNG',
    screenUnavailable: 'Preview unavailable',
    startGame: 'Start a game to see the screen',

    // Games
    systems: 'Systems',
    searchRoms: 'Search ROMs...',
    selectSystem: 'Select a system',
    selectSystemSub: 'to browse ROMs',
    noRoms: 'No ROMs in this system',
    launching: 'Launching...',
    launchError: 'Launch error',
    noSystemsFound: 'No ROMs found on SD card',

    // Files
    upload: '⬆ Upload',
    savesBackup: '⬇ Saves',
    dropToUpload: 'Drop files here\nto upload to',
    deleteConfirm: (name) => `Delete "${name}"?`,
    deleteSuccess: 'Deleted: ',
    uploadComplete: 'Upload complete',
    extract: 'Extract',
    emptyDir: 'Empty directory',

    // Config
    configFiles: 'Config Files',
    selectConfig: 'Select a file',
    backupNote: 'Auto-backup before every save',
    save: 'Save (Ctrl+S)',
    saving: 'Saving...',
    unsaved: 'Unsaved changes',
    savedMsg: '✓ Saved',
    notEditable: 'This file type cannot be edited',
  }
}

// Detect language from browser
function detectLang() {
  const saved = localStorage.getItem('miyoodeck_lang')
  if (saved && translations[saved]) return saved
  const nav = navigator.language || 'en'
  return nav.startsWith('fr') ? 'fr' : 'en'
}

import { writable, derived } from 'svelte/store'

export const lang = writable(detectLang())
export const t = derived(lang, ($lang) => translations[$lang])

export function setLang(l) {
  lang.set(l)
  localStorage.setItem('miyoodeck_lang', l)
}

export const availableLangs = [
  { code: 'fr', label: '🇫🇷 Français' },
  { code: 'en', label: '🇬🇧 English' },
]
