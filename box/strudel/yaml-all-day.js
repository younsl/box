// YAML ALL DAY
// Warm, emotional house/funk vibe
// @version 1.0
setcps(0.475) // 95 BPM - laid-back groove

// Chord progression
let chords = "<Dmaj9 Gmaj7 Em7 A7sus4>/2"

stack(
  // DRUMS - Classic house pattern
  s("bd:3*2").gain(0.9), // Kick
  s("~ cp").room(0.3).shape(0.2), // Clap
  s("[~ hh:2]*4").gain(0.5).pan(sine.range(-0.3, 0.3)), // Hi-hats
  s("oh:1").struct("~ ~ ~ [~ x]").gain(0.3), // Open hat
  
  // BASS - Deep and warm
  note("<d2 g2 e2 a2>/2")
    .struct("x(3,8)")
    .s("sawtooth")
    .lpf(400).lpq(10)
    .gain(0.8).room(0.1),
  
  // ELECTRIC PIANO - Main chords
  chords.voicings('lefthand')
    .s("gm_epiano1")
    .struct("~ [x ~ x ~]")
    .velocity(0.6)
    .room(0.5).delay(0.125)
    .gain(0.7),
  
  // PAD - Atmospheric strings
  chords.voicings()
    .add(12) // octave up
    .s("gm_string_ensemble_1")
    .lpf(perlin.range(800, 2000).slow(4))
    .pan(sine.range(-0.5, 0.5).slow(8))
    .room(0.6)
    .gain(0.3),
  
  // LEAD MELODY  
  note("<d5 ~ [e5 f#5] ~ g5 ~ [a5 g5] ~>/2")
    .s("sawtooth")
    .struct("<x ~ ~ x [~ x] ~ x ~>/2")
    .lpf(sine.range(500, 2000).slow(2)).lpq(15)
    .vowel("<a e i o u>".slow(4))
    .room(0.4).delay(0.25)
    .gain(0.5)
    .mask("<0 0 1 1>/8"), // Lead comes in halfway
  
  // RHYTHM GUITAR
  chords.voicings()
    .struct("~ [x!2 ~] ~ [~ x]")
    .s("gm_guitar_harmonics")
    .velocity(perlin.range(0.5, 0.8))
    .room(0.3).shape(0.3)
    .pan(-0.3)
    .gain(0.4)
    .sometimes(ply(2)),
  
  // VOCODER PAD - Texture
  note("<[d4,g4,b4] [g4,b4,d5] [e4,g4,b4] [a4,c#5,e5]>/2")
    .s("supersquare")
    .struct("x*8")
    .lpf(sine.range(200, 800).slow(16))
    .vowel("<o a e>".slow(8))
    .room(0.8).delay(0.5)
    .gain(0.15)
    .mask("<0 0 0 1>/16"),
  
  // PERCUSSION
  s("shaker:2").struct("~ x*2 ~ x").gain(0.25).pan(0.4),
  s("conga:3").struct("~ ~ [~ x] ~").gain(0.35).pan(-0.2),
  
  // ADDITIONAL ATMOSPHERE
  note("d5,a5")
    .s("gm_pad_warm")
    .struct("x ~ ~ ~")
    .lpf(1000)
    .room(0.9)
    .gain(0.2)
    .mask("<1 0 0 0>/8")
)
.late(0.002) // Humanization
