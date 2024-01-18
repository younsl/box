package org.springframework.samples.petclinic.system;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class LongProcessController {

    @RequestMapping("/40process")
    public String pause_40() throws InterruptedException {
        Thread.sleep(40000);
        return "40 sec Process finished";
    }

    @RequestMapping("/20process")
    public String pause_20() throws InterruptedException {
        Thread.sleep(20000);
        return "20 sec Process finished";
    }

}