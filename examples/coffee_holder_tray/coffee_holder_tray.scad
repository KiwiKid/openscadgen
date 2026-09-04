

	include <BOSL2/std.scad>;
    include <BOSL2/hooks.scad>; 
    
    trayWallThickness = 2;
    trayThickness = 3;
    traySize = [180,50,trayThickness];
    
    coffeeCupShift = 60;
    coffeeCupTopRadius = 20;
    
    coffeeCupBottomRadius = 25;
    coffeeCupHeight = 50;
    coffeeCupWallDepth = 2;
    coffeeCupSink = 10;
    
    trayDown = 10;
    insideTraySize = [30,30,30];
    
    
    
    module coffee_cups(mode=mode){
         left(coffeeCupShift)
    coffee_cup(mode=mode);
    
             right(coffeeCupShift)
    coffee_cup(mode=mode);
    }
    
    module coffee_cup(mode="middleCutout"){
    difference(){
        cyl(r1=coffeeCupTopRadius, r2=coffeeCupBottomRadius, h=coffeeCupHeight, anchor=BOT);
        if(mode == "middleCutout"){
        up(coffeeCupWallDepth)
        cyl(r1=coffeeCupTopRadius-coffeeCupWallDepth, r2=coffeeCupBottomRadius-coffeeCupWallDepth, h=coffeeCupHeight, anchor=BOT);
        }

    }    
    }
    
    module middle_tray(mode="middleCutout"){
    difference(){
        cuboid(insideTraySize, anchor=BOT);
         if(mode == "middleCutout"){
        up(trayWallThickness)

            cuboid(insideTraySize-[1,0,0], anchor=BOT);
          }
        
    
    }
    }

module coffee_holder_tray(){
    difference(){
	cube(traySize, center=true);
    down(coffeeCupSink){
    middle_tray(mode="noCutout");

    coffee_cups(mode="noCutout");
    }
    }
    
    down(coffeeCupSink){
        coffee_cups(mode="middleCutout");
       middle_tray();
       }
        
}

module cup_holder_tray_hook(){
    
ring_hook([50, 10], 25, 25, ir=20);
}

coffee_holder_tray();

left(100)
cup_holder_tray_hook();
