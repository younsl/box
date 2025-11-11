use std::fs::OpenOptions;
use std::io::{self, Write};

/// Write a line directly to /dev/tty for interactive output
/// This bypasses stdout/stderr so kubectl logs won't capture REPL output
/// If /dev/tty is not available (non-interactive), silently ignore
pub fn tty_writeln(text: &str) -> io::Result<()> {
    match OpenOptions::new().write(true).open("/dev/tty") {
        Ok(mut tty) => {
            writeln!(tty, "{}", text)?;
            tty.flush()?;
            Ok(())
        }
        Err(_) => {
            // /dev/tty not available (non-interactive environment)
            // Silently ignore - this is expected behavior
            Ok(())
        }
    }
}

/// Write formatted output to /dev/tty
#[allow(dead_code)]
pub fn tty_write(text: &str) -> io::Result<()> {
    match OpenOptions::new().write(true).open("/dev/tty") {
        Ok(mut tty) => {
            write!(tty, "{}", text)?;
            tty.flush()?;
            Ok(())
        }
        Err(_) => {
            // /dev/tty not available (non-interactive environment)
            // Silently ignore - this is expected behavior
            Ok(())
        }
    }
}
