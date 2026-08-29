

	include <BOSL2/std.scad>;
    include <BOSL2/hooks.scad>;
    
    trayWallThickness = 2;
    trayThickness = 3;
    traySize = [150,50,trayThickness];
    
    coffeeCupShift = 60;
    coffeeCupTopRadius = 20;
    
    coffeeCupBottomRadius = 25;
    coffeeCupHeight = 50;
    coffeeCupWallDepth = 2;
    coffeeCupSink = 10;
    
    insideTraySize = [10,10,10];
    
    
    
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
        cuboid(insideTraySize);
         if(mode == "middleCutout"){
        up(trayWallThickness)
        cuboid(insideTraySize);
          }
        
    
    }
    }

module coffee_holder_tray(){
    difference(){
	cube(traySize, center=true);
    #middle_tray();
    down(coffeeCupSink)
    coffee_cups(mode="noCutout");
    }
    
    down(coffeeCupSink)
        coffee_cups(mode="middleCutout");
       #middle_tray();
        
}

module cup_holder_tray_hook(){
    
ring_hook([50, 10], 25, 25, ir=20);
}

coffee_holder_tray();

left(100)
cup_holder_tray_hook();
