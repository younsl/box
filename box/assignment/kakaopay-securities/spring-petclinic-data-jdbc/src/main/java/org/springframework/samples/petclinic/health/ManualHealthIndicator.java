package org.springframework.samples.petclinic.health;

import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;

public interface ManualHealthIndicator extends HealthIndicator {

    void setHealth(Health health);
}