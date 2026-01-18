

	include <BOSL2/std.scad>;

	include <BOSL2/joiners.scad>;
	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;


isSizer = "true";
sizerLength = 200;
sizerCount = 5;


sizerHeight = 4;
onlySection = "3";

	module big_label(){
    
    name = "DIDDYS";
 text_angle = 10;
 base_width = !is_undef(base_width) ? base_width : 130 ;
base_size = [base_width, 30, 20];

text_size = 180;
text_height = 2;

include_base = "false";
include_connector = "false";
connector_size = [300, 3,3];





            difference(){
            union(){



          //  rotate([-text_angle,0,0])
          //  cylindrical_extrude(or=140, ir=110)
          rotate([0,0,180])
          linear_extrude(height = text_height)
            text(text=name, size=text_size, halign="center", valign="center", font="Baskerville");
            
            
            if(include_base =="true"){
                fwd(60)
                up(93)
                cuboid(base_size, rounding=2);
                }
             
            if(include_connector == "true"){
                // Middle cylinder (straight)
                fwd(20)
                up(text_height+1.5)
                left(150)
                rotate([90, 0, 90])  // Rotate so cylinder lies along X
                cylinder(h=connector_size[0], r=connector_size[1]/2, center=false);

                // Left angled cylinder
                fwd(25)
                up(text_height+1.5)
                rotate([0, 0, 50])
                right(10)
                back(100)
                rotate([90, 0, 0])
                cylinder(h=105, r=connector_size[2]/2, center=false);

                // Middle lower cylinder (optional duplicate? removing it)
                /*
                fwd(25)
                up(text_height+1.5)
                cylinder(h=connector_size[0], r=connector_size[1]/2, center=false);
                */

                // Right angled cylinder
                up(text_height+1.5)
                right(80)
                rotate([0, 0, -50])
                right(-8)
                back(50)
                
                rotate([90, 0, 0])
                cylinder(h=90, r=connector_size[2]/2, center=false);
            }
            
           
            
            }
	}
    }

    
    module sizerSection(){
    
    if(onlySection == str(sizerCount-$idx) || onlySection == "all"){
        difference(){

        
       diff("remove")
       down(sizerHeight/2-1)
       rotate([0,0,90])
        cuboid([50,sizerLength,sizerHeight], anchor=CENTER){
            attach(BACK) dovetail("male", slide=sizerHeight, width=25, height=25, radius=0.5, round=true);
            tag("remove")attach(FRONT) dovetail("female", slide=sizerHeight, width=25, height=25, radius=0.5, round=true);
            
            
            
        }
        
        move([8,12,0])
        rotate([0,0,180])
        text3d(str(sizerCount-$idx), h=30, size=30);
        }
        }
    }

    if(isSizer == "true"){
    
  difference(){
   
    up(3)
    xcopies(sizerLength, n=sizerCount){
    
        sizerSection();
    }
    #big_label();
    }
 
 } else{ 
    
 big_label();
 }
 
    
       




