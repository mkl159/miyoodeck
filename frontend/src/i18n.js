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
    logoutTip: 'Déconnexion',

    // Dashboard
    cpu: 'Processeur',
    ram: 'Mémoire',
    battery: 'Batterie',
    network: 'Réseau',
    uptime: 'Uptime',
    temp: 'Température',
    brightness: 'Luminosité',
    liveScreen: 'Écran en direct',
    refresh: 'Rafraîchir',
    savePng: '⬇ Capture PNG',
    screenUnavailable: 'Aperçu non disponible',
    startGame: 'Lancez un jeu pour voir l\'écran',
    power: 'Alimentation',
    reboot: 'Redémarrer',
    poweroff: 'Éteindre',
    rebootConfirm: 'Redémarrer la console maintenant ?',
    poweroffConfirm: 'Éteindre la console maintenant ?',

    // Games
    systems: 'Systèmes',
    searchRoms: 'Rechercher des ROMs...',
    searchAll: 'Rechercher dans tous les systèmes...',
    selectSystem: 'Sélectionnez un système',
    selectSystemSub: 'ou utilisez la recherche globale 🔍',
    noRoms: 'Aucune ROM dans ce système',
    launching: 'Lancement...',
    launchError: 'Erreur de lancement',
    noSystemsFound: 'Aucune ROM trouvée sur la carte SD',
    surpriseMe: '🎲 Surprends-moi',
    searchResults: 'Résultats',
    favorites: 'Favoris',
    noFavorites: 'Aucun favori — clique sur ★ pour en ajouter',
    smoothStream: 'Flux fluide',
    konami: 'Code Konami envoyé ! 🎉',
    quitGame: '⏹ Quitter le jeu',
    quitGameConfirm: 'Quitter le jeu en cours ? (la sauvegarde sera effectuée)',
    kbdControl: '⌨ Clavier',
    kbdHint: 'Flèches = croix · X=A · Z=B · S=X · A=Y · Q/E=L1/R1 · W/R=L2/R2 · Entrée=START · Maj=SELECT · M=MENU',
    logs: 'Logs serveur',
    showLogs: '📜 Logs',
    refreshLogs: '↻ Rafraîchir',

    // Files
    upload: '⬆ Envoyer',
    savesBackup: '⬇ Sauvegardes',
    dropToUpload: 'Déposez les fichiers ici\npour les envoyer vers',
    deleteConfirm: (name) => `Supprimer "${name}" ?`,
    deleteSuccess: 'Supprimé : ',
    uploadComplete: 'Envoi terminé',
    extract: 'Extraire',
    emptyDir: 'Dossier vide',
    newFolder: '📁+ Dossier',
    newFolderPrompt: 'Nom du nouveau dossier :',
    renameTip: 'Renommer',
    renamePrompt: 'Nouveau nom :',
    downloadTip: 'Télécharger',
    deleteTip: 'Supprimer',

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
    logoutTip: 'Log out',

    // Dashboard
    cpu: 'CPU',
    ram: 'RAM',
    battery: 'Battery',
    network: 'Network',
    uptime: 'Uptime',
    temp: 'Temperature',
    brightness: 'Brightness',
    liveScreen: 'Live Screen',
    refresh: 'Refresh',
    savePng: '⬇ Save PNG',
    screenUnavailable: 'Preview unavailable',
    startGame: 'Start a game to see the screen',
    power: 'Power',
    reboot: 'Reboot',
    poweroff: 'Power off',
    rebootConfirm: 'Reboot the console now?',
    poweroffConfirm: 'Power off the console now?',

    // Games
    systems: 'Systems',
    searchRoms: 'Search ROMs...',
    searchAll: 'Search across all systems...',
    selectSystem: 'Select a system',
    selectSystemSub: 'or use the global search 🔍',
    noRoms: 'No ROMs in this system',
    launching: 'Launching...',
    launchError: 'Launch error',
    noSystemsFound: 'No ROMs found on SD card',
    surpriseMe: '🎲 Surprise me',
    searchResults: 'Results',
    favorites: 'Favorites',
    noFavorites: 'No favorites yet — tap ★ to add some',
    smoothStream: 'Smooth stream',
    konami: 'Konami code sent! 🎉',
    quitGame: '⏹ Quit game',
    quitGameConfirm: 'Quit the running game? (it will save first)',
    kbdControl: '⌨ Keyboard',
    kbdHint: 'Arrows = D-pad · X=A · Z=B · S=X · A=Y · Q/E=L1/R1 · W/R=L2/R2 · Enter=START · Shift=SELECT · M=MENU',
    logs: 'Server logs',
    showLogs: '📜 Logs',
    refreshLogs: '↻ Refresh',

    // Files
    upload: '⬆ Upload',
    savesBackup: '⬇ Saves',
    dropToUpload: 'Drop files here\nto upload to',
    deleteConfirm: (name) => `Delete "${name}"?`,
    deleteSuccess: 'Deleted: ',
    uploadComplete: 'Upload complete',
    extract: 'Extract',
    emptyDir: 'Empty directory',
    newFolder: '📁+ Folder',
    newFolderPrompt: 'New folder name:',
    renameTip: 'Rename',
    renamePrompt: 'New name:',
    downloadTip: 'Download',
    deleteTip: 'Delete',

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
